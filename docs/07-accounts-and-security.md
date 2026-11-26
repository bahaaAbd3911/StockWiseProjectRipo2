# 7. Accounts and security

## 7.1 Sign-up

`POST /api/register`:

1. Validate the input.
2. Check the email is free **in this store**. The check ignores letter case.
3. Hash the password with bcrypt.
4. Create a **customer**. A client can never make itself admin.
5. Email a confirmation link.

Logging in works right away. Only **placing an order** needs a confirmed
email.

**Why confirm emails at all?** Without it anyone could sign up with someone
else's address, and order and debt emails would go to that person.

**Why only for ordering?** Customers can browse, fill a cart and save
favorites immediately; only the step where emails start going out needs a
real address. Admins are never blocked.

## 7.2 Email confirmation

```
register → account (email_verified = false) → email: <store>/verify-email?token=...
     │                                                │
can log in, browse, cart                  the page posts the token: POST /api/email/verify
     │                                                │
POST /api/orders → 403 "please confirm your email"   email_verified = true → checkout works
```

- The token is 32 random bytes; only its SHA-256 is stored. It is valid
  48 h and works once.
- "Send again" (`POST /api/email/verify/resend`, 3 per hour) deletes the
  old link.
- Every bad link gets the same message: "invalid or has expired" (400).
- Accounts that existed before the feature were marked confirmed by
  migration 000004, so no existing customer was suddenly blocked.
- `REQUIRE_EMAIL_VERIFICATION=false` turns the checkout rule off.

## 7.3 Login

`POST /api/login` returns `{user, token, expires_at}`.

- Unknown emails and wrong passwords get **the same answer at the same
  speed**: a dummy bcrypt comparison runs for unknown emails. Nobody can use
  the login form to find out who is a customer.
- The token is an HS256 JWT signed with `JWT_SECRET`. It carries the user
  id, the store (`sid`), a version (`ver`), the issuer and the expiry
  (`JWT_TTL`, default 24 h).
- Parsing **pins** the algorithm to HS256, which blocks the `alg: none`
  trick, and requires issuer and expiry.

On **every** protected request, `Authenticate` asks the database for the
user's current `is_active`, `role`, `token_version` and email status. So:

- a deactivated account is locked out at once;
- a role always comes from the database, never from the token alone;
- a revoked token stops working at once.

## 7.4 Revocation: "log out everywhere"

`users.token_version` is copied into each token as `ver`. Raising it kills
every older token:

| Action | Effect |
|---|---|
| change password (`PUT /api/profile/password`) | other devices logged out; the response carries a **new token** for this device |
| reset password | every device logged out |
| `POST /api/logout-all` | every device, including this one |
| admin changes a user's password or role, or deactivates them | that user is logged out |

A normal log-out on one device needs no server call: the frontend deletes
its token.

## 7.5 Password reset

1. `POST /api/password/forgot {email}` **always** answers the same message,
   at the same speed. The lookup and the email happen in the background
   after the reply.
2. If the account exists and is active, the API creates a random token,
   stores its hash with a 30-minute expiry (`PASSWORD_RESET_TTL`) and emails
   `<store>/reset-password?token=...`.
3. `POST /api/password/reset {token, new_password}`:
   - checks the token: known, unused, not expired;
   - sets the password;
   - raises `token_version`;
   - marks the token used;
   - deletes the user's other links.

Further rules:

- Only the latest link works: a new request deletes older unused links.
- Every bad token gets one message (400), so nothing is revealed.
- The requester's IP is stored with the token for auditing.

## 7.6 Rate limits

Every limit is a **token bucket**: a key may make a burst of N requests, and
its budget refills steadily over the window.

| Limit | Key | Default | Applies to | When exceeded |
|---|---|---|---|---|
| General | client IP | 300 / minute | every request | 429 |
| Account endpoints | client IP | 10 / minute | login, register, password forgot/reset, email verify, platform login, **applying a coupon** | 429 |
| Failed logins | email + store | 5 / 15 minutes | wrong password or unknown email | 429 from **any** IP |
| Reset emails | email + store | 3 / hour | password forgot | **silent**: same reply, no email |

Why each limit is built this way:

- **Why a per-email limit too?** An attacker with many IPs gets past the
  per-IP limit; the per-email counter still stops password guessing against
  one account. A successful login resets it.
- **Why is the reset limit silent?** Answering "too many requests" for one
  email and not another would reveal which emails have accounts.
- **Why is applying a coupon limited?** So codes cannot be guessed by trying
  thousands of them.
- **The trade-off:** someone who knows a customer's email can lock that
  account's login for 15 minutes by failing on purpose. The customer can
  still reset the password. The numbers are in `config.DefaultRateLimits()`.
- **The client IP:** by default the TCP address. Behind a proxy, set
  `TRUST_PROXY=true` so the first `X-Forwarded-For` address is used. Never
  turn it on without a proxy: clients could send a fake header and get
  unlimited attempts.
- **The counters live in memory.** A restart clears them. With several API
  servers each counts separately; move them to Redis then (only the
  `ratelimit` package changes).

A 429 carries `Retry-After` (seconds) and the normal JSON error body.

## 7.7 Security headers

| Where | Headers |
|---|---|
| API responses (`middleware.SecurityHeaders`) | `X-Content-Type-Options: nosniff`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, HSTS in production |
| The React pages (`deploy/Caddyfile`) | a CSP allowing only the app's own scripts, Google Fonts and cdnjs for fonts and icons, and https images; `X-Frame-Options: DENY`; `nosniff`; `strict-origin-when-cross-origin` referrer; HSTS |

**Why both?** They protect different things:

- The API's headers make sure a JSON answer can never be run as a page.
- The page headers are what stops two attacks:
  - **clickjacking**: another site framing the admin panel to trick an
    admin into clicking;
  - **injected scripts**, which could read the login token from
    localStorage.

## 7.8 Platform admin logins

Platform admins are a separate table with tokens from a different issuer
(`ecommerce-platform`). They only work on `PLATFORM_ADMIN_HOST`, and that
host serves nothing else. See [15-platform-admin.md](15-platform-admin.md).
