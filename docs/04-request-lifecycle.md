# 4. One request, from click to database

Example: a customer of Ahmad's store adds a product to the cart.

## The journey

1. **Browser.** The React app calls
   `POST https://ahmad.myplatform.com/api/cart/items` with:
   - the login token in `Authorization: Bearer ...`;
   - the body `{"product_id": 7, "quantity": 1}`.
   It is the same address as the page, so the browser needs no CORS
   permission.

2. **Caddy.** It has a certificate for `ahmad.myplatform.com` and forwards
   the request to the API. It keeps the original host name and adds the
   visitor's IP in `X-Forwarded-For`.

3. **Global middleware**, which runs on every request, outermost first
   (`internal/middleware`):

   | Step | What it does |
   |---|---|
   | Recover | a crash (panic) in a handler becomes a 500; the server keeps running |
   | Logger | one log line per request: method, path, status, duration |
   | SecurityHeaders | `nosniff`, no framing, a strict CSP, HSTS in production |
   | RealIP | the client's IP; `X-Forwarded-For` is only trusted when `TRUST_PROXY=true` |
   | CORS | allows the configured origins and answers preflight requests (before the store check, so even a 404 for an unknown store is readable by the browser) |
   | Store | finds the store from the host name: `ahmad` → store #12 (unknown → 404, suspended → 403) |
   | RateLimit | the general per-IP budget; over it → 429 with `Retry-After` |

4. **Route.** `routes.go` says `POST /api/cart/items` is a `user` route, so
   `Authenticate` runs. It checks:
   - the token's signature, algorithm, issuer and expiry;
   - that **the token's store is #12**;
   - that the user exists in store #12, is active and has the role the token
     says;
   - that the token's version matches the database (not revoked).

5. **Handler** (`handlers/cart_handler.go`) reads the JSON (size-limited) and
   calls **one** service method.

6. **Service** (`services/cart_service.go`) validates the input: quantity
   1–1000, and a color is required if the product has colors. It then calls
   the repository.

7. **Repository** (`repository/cart_repository.go`) runs the SQL.

8. **Database connection.** Before the SQL runs, the connection wrapper sets
   `app.store_id = 12` (only if it was different). PostgreSQL's row-level
   security then lets this query see and change **only store #12's rows**. If
   product 7 belongs to another store, it simply is not found → 404.

9. **Response.** The whole updated cart comes back as JSON with live prices.
   The React mapper converts it, the cart context updates, and every
   component that shows the cart re-renders.

## How errors travel back

Every layer returns a **typed error** instead of writing a response:

```go
return apperr.NotFound("product not found")
```

One function, `httpx.WriteError`, turns the kind into a status and a body:

```json
{"error": "validation failed", "fields": {"quantity": "must be between 1 and 1000"}}
```

| Kind | Status |
|---|---|
| Invalid / Validation | 400 |
| Unauthorized | 401 |
| Forbidden | 403 |
| NotFound | 404 |
| Conflict | 409 |
| TooManyRequests | 429 + `Retry-After` |
| anything unexpected | 500, details only in the server log |

Database errors are translated too:

| Database error | Response |
|---|---|
| no rows | 404 |
| unique violation | 409 |
| foreign key | 404 |
| check or length | 400 |

**Why:**

- Code never compares error message strings.
- Every endpoint answers errors in the same shape, so the frontend handles
  them in one place (`api/client.js`).
- Internal details such as SQL never reach a browser.

## Background work (not tied to a request)

- **`notify.Worker`**: every few seconds it sends queued emails and retries
  failures; about every 10 minutes it queues debt reminders. It runs in
  "platform mode" because it serves every store (see
  [05-multi-tenancy.md](05-multi-tenancy.md)).
- **Account emails** (confirmation, password reset) are sent right after the
  response, so a slow mail server never slows the reply and the reply never
  reveals whether an account exists.
