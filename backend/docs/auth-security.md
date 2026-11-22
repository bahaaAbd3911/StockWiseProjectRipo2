# Account security: rate limiting, login revocation, password reset

Added in migration `000003_auth_security`. Apply it before starting the API:

```powershell
go run ./cmd/migrate up
```

- [1. Rate limiting](#1-rate-limiting)
- [2. Login revocation (log out everywhere)](#2-login-revocation-log-out-everywhere)
- [3. Password reset](#3-password-reset)
- [4. Configuration](#4-configuration)
- [5. Endpoints](#5-endpoints)
- [6. For the frontend](#6-for-the-frontend)
- [7. Files](#7-files)
- [8. Tests](#8-tests)
- [9. Limits and next steps](#9-limits-and-next-steps)

---

## 1. Rate limiting

Four limits protect the API. Each is a **token bucket**: a key may make a
burst of N requests, and its budget refills steadily until it is full again
after the time window.

| Limit | Key | Default | Applies to | When exceeded |
|---|---|---|---|---|
| General | client IP | 300 / minute | every request | 429 |
| Account endpoints | client IP | 10 / minute | `POST /api/login`, `/api/register`, `/api/password/forgot`, `/api/password/reset` | 429 |
| Failed logins | email | 5 / 15 minutes | wrong password or unknown email | 429: the account refuses logins for a while, from **any** IP |
| Reset emails | email | 3 / hour | `POST /api/password/forgot` | silently: same reply, but no email is sent |

A 429 response looks like every other error and carries `Retry-After` (seconds):

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 6
Content-Type: application/json

{"error":"too many requests, please try again later"}
```

Details:
- **Why a per-email limit too?** An attacker using many IPs gets past the
  per-IP limit; the per-email counter still stops password guessing against
  one account. A successful login resets that counter.
- **Trade-off:** someone who knows a customer's email can lock that account's
  login for up to 15 minutes by failing on purpose. The customer can still
  reset their password. This is the usual trade-off; raise the numbers in
  `config.DefaultRateLimits` if it becomes a problem.
- **The reset-email limit is silent** on purpose: replying "too many requests"
  for one email but not another would reveal which emails are accounts.
- **Client IP:** by default the IP of the TCP connection. Behind a reverse
  proxy every request seems to come from the proxy, so set `TRUST_PROXY=true`
  there: the first address of `X-Forwarded-For` (or `X-Real-IP`) is used.
  **Never** turn it on without a proxy, because clients could then send a
  fake header and get unlimited attempts.
- CORS preflight (`OPTIONS`) requests are not counted, and 429 responses
  still carry CORS headers so the browser can read them.
- **Memory:** the counters live in the server process. Restarting clears them,
  and with several servers behind a load balancer each counts separately. For
  more than one instance, move the buckets to Redis (the `ratelimit` package
  is the only place to change).

## 2. Login revocation (log out everywhere)

Every login token (JWT) now carries a version number (`"ver"` claim) copied
from the new column `users.token_version`. On **every** protected request the
authentication middleware compares the token's version with the database. If
the column has moved on, the token is refused:

```json
401 {"error": "this login has been revoked; please log in again"}
```

The version goes up, revoking every token issued before, when:

| Action | Effect |
|---|---|
| `PUT /api/profile/password` (change password) | all other devices are logged out; the response contains a **new token** for the current device |
| `POST /api/password/reset` (forgot password) | every device is logged out; the user logs in with the new password |
| `POST /api/logout-all` | every device, including this one, is logged out |

A normal "log out" on one device needs no server call: the frontend deletes
its token. Tokens issued before this change have no `ver` claim, which counts
as version 0 (the starting value), so nobody is logged out by the upgrade.

Deactivated accounts were already refused on every request (`is_active` is
checked too).

## 3. Password reset

### Flow

```
Customer                     Frontend                          API
   │ "Forgot password"  ───►  POST /api/password/forgot {email}
   │                                                       ├─ always 200, same message
   │                                                       └─ in the background:
   │                                                            account exists & active?
   │                                                            new random token (32 bytes)
   │                                                            store SHA-256(token), expires in 30 min
   │ ◄──────────── email with link  http://.../reset-password?token=XYZ
   │ opens the link  ───►  page reads ?token, asks for a new password
   │                       POST /api/password/reset {token, new_password}
   │                                                       ├─ token valid, unused, not expired?
   │                                                       ├─ set password, token_version + 1
   │                                                       └─ mark token used, delete other links
   │ ◄──────────── 200 "password has been reset; please log in"
```

### Security properties

- **No account enumeration.** `/api/password/forgot` gives the same answer,
  at the same speed, for unknown, deactivated and real emails. The lookup and
  the email happen in the background after the reply.
- **Tokens are unguessable**: 32 random bytes, URL-safe base64.
- **Only a hash is stored** (`password_reset_tokens.token_hash` = SHA-256).
  Someone who reads the database cannot use the links.
- **Single use, short life**: a token works once and only until `expires_at`
  (default 30 minutes).
- **Only the latest link works**: a new request deletes the user's older
  unused links.
- **One error for every bad token** (unknown, used, expired, account
  deactivated): "this password reset link is invalid or has expired", 400.
- After a reset **all sessions are revoked** (token_version + 1).
- The requester's IP is stored with the token for auditing.

### Emails

The email is plain text in English and Arabic and contains the link
`PASSWORD_RESET_URL?token=...`.

| `MAIL_DRIVER` | What happens |
|---|---|
| `log` (default) | the email, including the link, is printed in the server log. Use this in development: copy the link from the terminal |
| `smtp` | sent through your provider (`SMTP_HOST`, `SMTP_PORT` 587 with STARTTLS, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`) |

Any SMTP service works: Brevo, Mailgun, Amazon SES, Postmark, or Gmail with an
app password for testing. For good delivery, send from your own domain with
SPF/DKIM set up at the provider.

The frontend needs a page at `PASSWORD_RESET_URL` (default
`http://localhost:5173/reset-password`) that reads `token` from the query
string and posts it with the new password.

## 4. Configuration

| Variable | Default | Meaning |
|---|---|---|
| `RATE_LIMIT_ENABLED` | `true` | switch all limits on/off |
| `TRUST_PROXY` | `false` | read the client IP from `X-Forwarded-For` / `X-Real-IP`; only behind a proxy |
| `PASSWORD_RESET_URL` | `http://localhost:5173/reset-password` | frontend page for the link |
| `PASSWORD_RESET_TTL` | `30m` | link lifetime (1m - 24h) |
| `MAIL_DRIVER` | `log` | `log` or `smtp` |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM` | -, `587`, -, -, - | SMTP settings (`SMTP_HOST` and `SMTP_FROM` required for `smtp`) |

The numbers of the limits are in `config.DefaultRateLimits()`.

## 5. Endpoints

| Method | Path | Access | Body | Response |
|---|---|---|---|---|
| POST | `/api/password/forgot` | Public (rate-limited) | `{email}` | 200 `{message}` - always the same |
| POST | `/api/password/reset` | Public (rate-limited) | `{token, new_password}` | 200 `{message}`; 400 invalid/expired token or weak password |
| POST | `/api/logout-all` | User | - | 200 `{message}` - every token of the user stops working |
| PUT | `/api/profile/password` | User | `{current_password, new_password}` | **changed**: 200 `{message, token, expires_at}` |
| POST | `/api/login`, `/api/register` | Public | unchanged | now rate-limited; login may answer 429 |

## 6. For the frontend

- **Handle 429** everywhere: show the message and, if useful, wait
  `Retry-After` seconds before allowing another try.
- **Handle 401 "revoked"** like an expired token: delete the stored token and
  send the user to the login page.
- **After changing the password**, replace the stored token with the `token`
  from the response, or the user will be logged out on the next request.
- **Log out** = delete the token locally. Offer "log out on all devices" in
  account settings → `POST /api/logout-all`, then delete the token.
- **Forgot password page**: email field → `/api/password/forgot` → always
  show "If an account exists, we sent a link."
- **Reset page** at `PASSWORD_RESET_URL`: read `token` from the URL → new
  password form → `/api/password/reset` → on success go to login.

## 7. Files

**New**

| File | Purpose |
|---|---|
| `migrations/000003_auth_security.up.sql` / `.down.sql` | `users.token_version`, `password_reset_tokens` |
| `internal/ratelimit/ratelimit.go` | token-bucket limiter (`Allow`, `Check`, `Fail`, `Reset`), memory clean-up |
| `internal/mail/mail.go` | `Mailer` interface, `LogMailer`, `SMTPMailer` |
| `internal/middleware/ratelimit.go` | `RealIP` (client IP into the context), `RateLimit` (per-IP 429) |
| `internal/repository/password_reset_repository.go` | store reset tokens, use one (password + revoke in one transaction) |
| `tests/auth_test.go` | end-to-end tests below |

**Changed**

| File | Change |
|---|---|
| `internal/auth/auth.go` | `ver` claim; `Issue(userID, version)`, `Parse` returns the version; `AccountInfo` |
| `internal/middleware/auth.go` | rejects tokens whose version is outdated |
| `internal/repository/user_repository.go` | reads `token_version`; `UpdatePassword` bumps it; `RevokeTokens` |
| `internal/services/user_service.go` | failed-login lock, new token after a password change, log out everywhere, reset request / reset |
| `internal/handlers/user_handler.go` | `LogoutEverywhere`, `ForgotPassword`, `ResetPassword` |
| `internal/apperr`, `internal/httpx` | 429 kind with `Retry-After`; client IP helpers |
| `internal/config/config.go` | settings above |
| `internal/routes/routes.go` | limiters, new routes, middleware order; `New` now takes a mailer and returns an `App` |
| `cmd/api/main.go` | builds the mailer; waits for queued emails on shutdown |

## 8. Tests

- **Unit:** token bucket (burst, refill, independent keys, check/fail/reset,
  clean-up, disabled), the token version in the JWT, the middleware refusing
  an old version, the client IP only taken from headers behind a trusted proxy.
- **`TestLoginRevocation`:** changing the password logs out the other devices
  with a "revoked" message and returns a working new token; log-out-everywhere
  revokes all tokens; logging in again works.
- **`TestPasswordReset`:** same reply for known and unknown emails, no email
  for unknown or deactivated accounts, only the hash stored, a new request
  invalidates the old link, bad/weak/used/expired tokens refused, all sessions
  revoked after a reset, old password stops working.
- **`TestRateLimits`** (second server with limits on and proxy trust, so it
  can simulate IPs): 429 + `Retry-After` after the per-IP budget, other IPs
  unaffected, normal pages unaffected, account lock after failed logins from
  different IPs, a successful login resets the counter, reset emails capped
  silently.

## 9. Limits and next steps

- Counters are in memory; use Redis when running more than one API instance.
- No CAPTCHA: consider one on login/register if bots become a problem.
- Email confirmation at sign-up now exists (same token pattern); see
  [GUIDE.md §6.3](GUIDE.md#63-email-confirmation-at-sign-up).
- Reset emails are sent directly by the API process (not through the email
  queue, which would store the secret link in the database). If the API
  restarts in the middle, the user simply asks again.
