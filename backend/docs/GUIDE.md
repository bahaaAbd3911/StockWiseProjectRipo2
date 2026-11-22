# StockWise backend - the complete guide

This guide explains the whole backend: what it does, how it is built, and the
logic behind every part, from the first rewrite to today. Three companion
documents go deeper into single topics:

- [offers-bundles-debts.md](offers-bundles-debts.md) - promotions, bundles and pay-later in detail
- [auth-security.md](auth-security.md) - rate limiting, login revocation, password reset in detail
- [production.md](production.md) - step-by-step deployment checklist

## Contents

1. [What the backend does](#1-what-the-backend-does)
2. [How we got here](#2-how-we-got-here)
3. [How a request travels through the code](#3-how-a-request-travels-through-the-code)
4. [Every folder and file](#4-every-folder-and-file)
5. [Ideas used everywhere](#5-ideas-used-everywhere)
6. [The features and their logic](#6-the-features-and-their-logic)
7. [Database and migrations](#7-database-and-migrations)
8. [Configuration and production mode](#8-configuration-and-production-mode)
9. [HTTPS](#9-https)
10. [Packaging and deployment](#10-packaging-and-deployment)
11. [Tests](#11-tests)
12. [API reference](#12-api-reference)
13. [What the frontend must do](#13-what-the-frontend-must-do)
14. [Limits and next steps](#14-limits-and-next-steps)

---

## 1. What the backend does

A REST API (JSON over HTTP) for an online shop with English and Arabic
content. Written in Go, data in PostgreSQL.

| Area | What customers and admins can do |
|---|---|
| Accounts | register, confirm email, log in, edit profile and profile picture, change or reset the password, log out everywhere, deactivate |
| Catalog | products with Arabic names, categories, colors, images (uploaded or by URL), search, filters, sorting |
| Promotions | timed **offers** (percentage or fixed, on products or a whole category) and **bundles** (several products for one price) |
| Shopping | cart (products and bundles), favorites, saved addresses |
| Orders | checkout (cash on delivery or **pay later**), tracking, cancelling, admin status workflow |
| Debts | credit limits, an append-only ledger of charges and payments, statements, reminders |
| Reviews | ratings and comments, "verified purchase", helpful votes |
| Waiting list | ask to be told when an out-of-stock product comes back |
| Emails | confirmation, password reset, order received / status changes, payments, debt reminders, back in stock |
| Protection | rate limits, revocable logins, production safety checks, HTTPS, security headers |

## 2. How we got here

| Step | Commit | What changed |
|---|---|---|
| Original code | `2a38421` .. `73527d4` | first backend: products, cart, orders, order statuses, reviews |
| **1. Rewrite** | `7bd6c21` | rebuilt against the real schema: layers (handlers → services → repository), typed errors, exact money (decimal, never float), validation with per-field messages, row locks at checkout, migration tool, unit + end-to-end tests |
| **2. Offers, bundles, debts** | `33d9fcf` | migration 000002: timed offers with "one best price", bundles sold as one product, pay-later with credit limits and a debt ledger |
| **3. Account security** | `bd86736` | migration 000003: rate limits, login revocation (token versions), password reset by email |
| **4. Production** | *current work* | production mode with safety checks, HTTPS (3 ways), security headers, Docker image + Compose + Caddy, email verification, notification emails with a durable queue, debt reminders, image uploads (disk or S3) - migration 000004 |

## 3. How a request travels through the code

```
Browser / app
   │  HTTPS
   ▼
Caddy or the hosting platform (HTTPS)          ── or the API itself with TLS_MODE=auto/files
   │
   ▼  global middleware (every request), outermost first
   │   Recover          a panic becomes a 500, the server keeps running
   │   Logger           one log line per request (method, path, status, time)
   │   SecurityHeaders  nosniff, no framing, CSP, HSTS in production
   │   RealIP           find the client's IP (X-Forwarded-For only behind a trusted proxy)
   │   CORS             allow the frontend's origin, answer preflight requests
   │   RateLimit        general per-IP budget → 429
   ▼
routes.go: which handler, and which access level
   │   public | sensitive (+ strict per-IP limit) | user (login) | verified (login + confirmed email) | admin
   ▼
middleware/auth.go: Authenticate → RequireVerifiedEmail / RequireAdmin
   ▼
handlers/    read the URL, query string, JSON or upload; call ONE service; write JSON
   ▼
services/    validate, apply the business rules, combine repositories, queue emails
   ▼
repository/  SQL, transactions, row locks
   ▼
PostgreSQL   constraints are the last safety net

Background (not tied to a request):
   notify.Worker   delivers queued emails, retries failures, queues debt reminders
   UserService     sends confirmation / reset emails right after the response
```

**Errors flow back up the same path.** A repository or service returns a
typed error such as `apperr.NotFound("order not found")`. The handler simply
returns it, and `httpx.WriteError` turns the kind into a status code and a
JSON body. Unknown errors are logged with details and the client only sees
`internal server error`, so SQL or internal messages never leak.

| Layer | May use | Must not contain |
|---|---|---|
| handlers | HTTP, JSON, services | SQL, business rules |
| services | models, validators, repositories, notify, storage | HTTP status codes, SQL |
| repository | SQL, models | HTTP |

## 4. Every folder and file

```
backend/
├── cmd/
│   ├── api/main.go            the server program
│   └── migrate/main.go        the migration program (status / up)
├── internal/                  all application code (Go forbids importing it from outside)
│   ├── apperr/                the error vocabulary (kinds → status codes)
│   ├── auth/                  login tokens (JWT) and the request's identity
│   ├── config/                every setting, defaults, validation, production checks
│   ├── database/              connection pool, transaction helper
│   ├── handlers/              HTTP in/out, one file per area
│   ├── httpx/                 JSON helpers, error writer, client IP
│   ├── images/                upload checks and metadata (GPS) removal
│   ├── mail/                  sending email: log driver and SMTP driver
│   ├── middleware/            auth, CORS, rate limit, security headers, logging, recover
│   ├── migrate/               the migration engine
│   ├── models/                data shapes and pure rules (status transitions)
│   ├── notify/                email texts, the email queue and its worker
│   ├── pricing/               the money math (best price, line totals)
│   ├── ratelimit/             in-memory token buckets
│   ├── repository/            all SQL
│   ├── routes/                builds the app and lists every endpoint
│   ├── server/                HTTP / HTTPS listeners, redirect, graceful stop
│   ├── services/              business rules
│   ├── storage/               uploaded files: local disk or S3
│   └── validators/            input normalising and checks
├── migrations/                numbered SQL files (000001 .. 000004, up and down)
├── tests/                     end-to-end tests against a throwaway database
├── deploy/Caddyfile           HTTPS reverse proxy configuration
├── docs/                      this guide and the deep dives
├── Dockerfile                 production image
├── docker-compose.yml         database + migrations + API + Caddy on one server
├── .env.example               development settings
└── .env.production.example    production settings
```

### `cmd/api/main.go`
Loads `.env` (optional) → `config.Load()` (production mode stops here if a
setting is unsafe, printing the whole list) → JSON logs in production →
start-up warnings → database pool → mailer → `routes.New()` → `server.Run()`.
Ctrl+C / SIGTERM: stop accepting connections, give running requests 10 s,
wait for account emails being sent, stop the email worker.
`api healthcheck` is a tiny sub-command used by Docker: it calls
`/api/health` on the running server and exits 0 or 1.

### `cmd/migrate/main.go` + `internal/migrate/`
`go run ./cmd/migrate status` (read-only) and `up`. Applies
`migrations/NNNNNN_name.up.sql` in order, each in its own transaction, and
records them in `schema_migrations`. A database whose baseline was created by
hand is recognised and only newer migrations run. The tests build their
database with the same engine.

### `internal/apperr/`
Kinds of expected errors: `Invalid` 400 (with per-field messages via
`Validation`), `Unauthorized` 401, `Forbidden` 403, `NotFound` 404, `Conflict`
409, `TooManyRequests` 429 (with Retry-After). `FromDB` translates database
errors: no rows → 404, unique violation → 409, foreign key → 404, check or
length → 400. Anything else is "unknown" → 500.

### `internal/auth/`
`TokenManager.Issue(userID, version)` signs an HS256 JWT (user id, token
version `ver`, issuer, expiry). `Parse` checks signature, algorithm (blocks
the `alg: none` trick), issuer and expiry. `Identity{UserID, Role,
EmailVerified}` is what handlers know about the caller.

### `internal/config/`
The only package that reads environment variables. Applies defaults, checks
formats, collects **every** problem before refusing to start, and in
production runs `checkProduction` (section 8).

### `internal/database/`
`Connect` opens the pool and pings. `WithTx(ctx, db, fn)` commits when `fn`
returns nil and rolls back otherwise - every multi-step write uses it.
`Querier` is satisfied by both the pool and a transaction.

### `internal/handlers/`
One file per area (`user`, `product`, `cart`, `order`, `review`, `favorite`,
`address`, `waiting_list`, `offer`, `bundle`, `debt`, `upload`, `health`).
Each function: read input → call one service → write JSON. They return
errors instead of writing them.

### `internal/httpx/`
`Handle` (adapts `func(w, r) error`), `WriteError`, `JSON`, `Message`,
`DecodeJSON` (1 MiB limit, clear messages for empty body / bad JSON / wrong
types), `PathID`, `Pagination`, and the client-IP helpers.

### `internal/images/`
`Inspect(data)` accepts JPEG, PNG and WebP only, decided from the bytes.
Reads only the header (size check, no full decoding, so a "decompression
bomb" cannot eat memory) and strips metadata: GPS/EXIF, XMP, IPTC, comments,
PNG text. JPEG orientation is kept in a minimal new EXIF block.

### `internal/mail/`
`Mailer` interface. `LogMailer` prints emails to the log (development: copy
links from the terminal). `SMTPMailer` sends through any SMTP provider with
STARTTLS and header-injection checks.

### `internal/middleware/`
- `auth.go`: `Authenticate` (token → database check of `is_active`, current
  role, token version, email verified), `RequireAdmin`, `RequireVerifiedEmail`.
- `ratelimit.go`: `RealIP`, `RateLimit`.
- `security.go`: `SecurityHeaders`.
- `cors.go`: allowed origins, preflight answers.
- `logging.go`: `Logger`, `Recover`.

### `internal/models/`
Plain structs for requests and responses, and pure rules that need no
database: `CanTransitionOrder`, `CanTransitionWaiting`, offer status, payment
methods. Money is `decimal.Decimal` everywhere, sent as JSON numbers.

### `internal/notify/`
- `templates.go`: every customer email, English then Arabic, with the shop
  name, links to the frontend, and amounts with the currency.
- `outbox.go`: `Outbox.Enqueue` stores an email in `email_outbox`; `Worker`
  delivers it, retries failures with growing delays, and queues debt
  reminders (section 6.14).

### `internal/pricing/`
The only place prices are decided: `DiscountedPrice`, `OfferPrice`,
`BestPrice` (single best of the product discount and live offers, never
stacked), `LineTotal`. Product page, cart and checkout always agree.

### `internal/ratelimit/`
Token buckets in memory: `Allow`, plus `Check` / `Fail` / `Reset` for
counting only failures (logins). Unused buckets are swept every minute.

### `internal/repository/`
All SQL, one file per area. Highlights: `cart_repository.go` has
`cartLine.problem()`, the single availability rule used by the cart view and
checkout; `order_repository.go` has the checkout transaction;
`offer_repository.go` has `offerAppliesSQL`, the single definition of "this
offer applies now"; `debt_repository.go` has the ledger helpers;
`email_verification_repository.go` and `password_reset_repository.go` handle
emailed tokens.

### `internal/routes/`
`routes.New` creates storage, the email queue and worker, every repository,
service and handler, registers every URL with its access level, wraps the
global middleware, and starts the worker. Reading it shows the whole API.

### `internal/server/`
`Run` starts plain HTTP, or HTTPS plus an HTTP → HTTPS redirect (section 9),
and stops everything gracefully.

### `internal/services/`
Business rules per area. They validate, decide, call repositories, and queue
emails after successful actions (an email problem never fails the action).

### `internal/storage/`
`Storage` interface with `Local` (files on disk, served at `/uploads/...`)
and `S3` (any S3-compatible bucket). Random file names only.

### `internal/validators/`
`Checker` collects every field problem at once. `rules.go` and `features.go`
have one `Validate...` per request: first **normalise** (trim, lower-case
emails, empty → NULL), then check against the schema's limits (lengths in
characters, money ≤ 99,999,999.99 with 2 decimals, phone, email, URL, hex
color).

## 5. Ideas used everywhere

**Typed errors, not strings.** Code never compares error messages; the kind
decides the status code. Validation errors carry `fields`:
`{"error": "validation failed", "fields": {"email": "is required"}}`.

**Exact money.** `numeric(10,2)` in the database, `decimal.Decimal` in Go. A
float would turn 0.1 + 0.2 into 0.30000000000000004.

**All-or-nothing writes.** Anything that touches more than one row runs in
`WithTx`. If any step fails, nothing is saved.

**Row locks against races.** `SELECT ... FOR UPDATE` makes concurrent requests
wait their turn where it matters: the last unit of stock, credit limits, the
one default address, the one primary image, reset and confirmation tokens,
the email queue.

**Not yours = not found.** Every query on personal data has
`WHERE user_id = $n`. Asking for someone else's order returns 404, as if it
did not exist.

**Soft deletes.** Products, bundles and accounts are deactivated, never
deleted, because orders reference them. Orders keep **copies** of names,
prices, colors and the address, so history never changes.

**One rule, one place.** Price → `pricing`; availability → `cartLine.problem`
and `bundleProblem`; offer is live → `offerAppliesSQL`; order status moves →
`CanTransitionOrder`. Every screen and checkout use the same code.

**Secrets are hashed.** Passwords with bcrypt; reset and confirmation tokens
with SHA-256. Someone reading the database cannot log in or use a link.

## 6. The features and their logic

### 6.1 Accounts and login
- **Register** (`POST /api/register`): validates, checks the email is free
  (case-insensitive), hashes the password with bcrypt, creates a *customer*
  (a client can never make itself admin), emails a confirmation link.
- **Login**: same response time for unknown emails as for wrong passwords (a
  dummy bcrypt comparison), one generic error, so nobody can find out which
  emails have accounts. Returns `{user, token, expires_at}`.
- **Every protected request**: the token is verified, then the database is
  asked for the user's *current* `is_active`, `role`, `token_version` and
  email status. A deactivated account is locked out at once, and a role can
  never come from the client.
- **First admin**: `UPDATE users SET role = 'admin' WHERE email = '...';`

### 6.2 Account security (summary - full details in [auth-security.md](auth-security.md))
- **Rate limits**: 300 requests/min per IP overall; 10/min per IP on login,
  register, password and email-confirmation endpoints; 5 failed logins per
  15 min per email (from any IP); 3 reset emails per hour per email (silent).
  429 responses carry `Retry-After`.
- **Revocation**: tokens carry `ver`; changing the password, resetting it or
  `POST /api/logout-all` raises `users.token_version`, which kills older
  tokens. A password change returns a fresh token for the current device.
- **Password reset**: always the same reply; random 32-byte token emailed,
  only its hash stored; 30 min, single use, newest link only; all sessions
  revoked afterwards.

### 6.3 Email confirmation at sign-up
Why: without it anyone can sign up with someone else's address, and order or
debt emails would go to the wrong person.

```
register ──► account created (email_verified = false) ──► email with link FRONTEND/verify-email?token=...
                     │                                              │
           can log in, browse, cart,                    frontend page posts the token:
           favorites, addresses                         POST /api/email/verify {token}
                     │                                              │
           POST /api/orders → 403                        email_verified = true
           "please confirm your email"                   checkout works (same login, no new token)
```

- Token: 32 random bytes, SHA-256 hash stored in
  `email_verification_tokens`, valid 48 h (`EMAIL_VERIFY_TTL`), single use;
  asking again (`POST /api/email/verify/resend`, 3 per hour) deletes the old
  link.
- Only **checkout** requires a confirmed email (`REQUIRE_EMAIL_VERIFICATION`,
  default on). Admins are never blocked.
- Accounts that existed before this feature were marked confirmed by
  migration 000004, so no current customer is suddenly blocked.
- One error for every bad link ("invalid or has expired"), 400.

### 6.4 Catalog
- Products: English + Arabic name, description and category, price, own
  discount percentage, stock, `is_active`. Computed fields: `final_price`
  (best price including live offers), `in_stock`, `primary_image_url`,
  `rating_average`, `review_count`, `offer` (the badge).
- List: `?category= &search= &sort=newest|price_asc|price_desc|name
  &in_stock=true &on_offer=true`, paginated. Sorts are whitelisted (no SQL
  injection through `sort`).
- Images: the first is automatically primary; exactly one primary (enforced
  by a unique index and a row lock); deleting the primary promotes the next.
- Colors: name, Arabic name, `#RRGGBB`; unique per product. Products with
  colors require a color in the cart.
- Delete = deactivate.

### 6.5 Image uploads
Two ways to give a product an image: by URL (`POST /api/products/{id}/images`)
or by uploading the file (`POST /api/admin/products/{id}/images/upload`).
Customers can upload a profile picture (`POST /api/profile/image`).

```
multipart upload ─► size cap (MAX_UPLOAD_MB, read straight from the stream, never to a temp file)
                ─► images.Inspect: real JPEG/PNG/WebP? sensible size? strip GPS & metadata
                ─► storage.Save: random name (32 hex + extension)
                     local: data/uploads/<name>  →  PUBLIC_URL/uploads/<name>
                     s3:    bucket/uploads/<name> →  S3_PUBLIC_URL/uploads/<name>
                ─► save the URL in product_images / users.profile_image_url
                   (if that fails, the file is deleted again)
```

- Why check the bytes: the file name and the browser's content type are
  chosen by the client. A "picture" that is really HTML could run scripts if
  served from our domain.
- Why strip metadata: phone photos record the GPS position where they were
  taken - often the customer's home.
- Serving (local): only names the API created, with `Cache-Control: immutable`
  (names never change), `nosniff`, and no directory listings.
- Deleting a product image or replacing / removing a profile picture also
  deletes the file. URLs that were typed in (not uploaded) are left alone.
- Local storage needs a disk that survives deploys (the Compose file uses a
  Docker volume). On platforms with a temporary disk, use S3 / Cloudflare R2.

### 6.6 Pricing and offers (details in [offers-bundles-debts.md](offers-bundles-debts.md))
- An offer: percentage or fixed amount, for listed products or a whole
  category, with a start and end time and an on/off switch.
- **One best price, no stacking**: the customer gets the lowest of the
  product's own discount and each live offer - never two discounts at once.
  On a tie, the product's own discount wins.
- Orders store which offer set the price and its name.

### 6.7 Bundles
- Several products (with fixed colors) sold together for one price, optionally
  within a time window. Shown with `original_total`, `savings`,
  `available_stock` (how many complete bundles the stock allows).
- In the cart and at checkout, a bundle's products count against stock
  together with the same products bought on their own.
- Cancelling gives back the stock of the products inside; delivering marks
  reviews of those products as verified purchases.

### 6.8 Cart
- One line per product (a second color is refused with a hint), bundles as
  their own lines. Every cart response shows live prices, line totals,
  `is_available` and `unavailable_reason`, and the subtotal.
- Availability is checked with the same rule that checkout uses, so the cart
  never promises what checkout will refuse.

### 6.9 Checkout and orders
Checkout is **one transaction**:
1. Address: the given one (must be yours) or your default.
2. Pay later: lock your user row (two checkouts cannot spend the same credit).
3. Lock your cart (no double checkout).
4. Lock every product the cart needs, in id order (no overselling, no deadlocks).
5. Check every line and the combined quantity per product; report all problems in one 409.
6. Prices from `pricing.BestPrice` / bundle prices.
7. Pay later: balance + total must fit the credit limit.
8. Insert the order with copies of the address, names, prices, colors, offer; reduce stock; add a debt `charge` for pay later.
9. Empty the cart, commit. Then queue the "order received" email.

The test `TestLastUnitCanOnlyBeSoldOnce` fires two checkouts at once for the
last unit: exactly one succeeds.

Life cycle (`models.CanTransitionOrder`; anything else → 409):

```
pending → confirmed → processing → shipped → delivered
   └──────────┴────────────┴──→ cancelled   (stock back, debt reversed)
```

- Customers may cancel while `pending`; admins until `processing`.
- Delivering: cash orders become `paid`, or `deferred` + a debt charge when
  the admin sends `payment_collected: false`; pay-later orders stay
  `deferred`; reviews of the products become verified purchases.
- The customer is emailed at every step (section 6.14).

### 6.10 Customer debts (pay later)
- An admin sets a **credit limit** per customer (0 = pay later off).
- The ledger `customer_debt_transactions` is **append-only**: `charge`
  (pay-later order, or cash order delivered unpaid), `payment`, `reversal`
  (automatic when a charged order is cancelled), `adjustment` (+ or -, note
  required). Balance = sum of the ledger; nothing is ever edited, so the full
  history stays visible.
- Statements show each row with the running balance.
- Emails: payment received (amount + remaining balance), balance corrected
  (amount, reason, new balance), and periodic reminders (section 6.14).

### 6.11 Reviews
One review per customer per product; rating 1-5, comment in English and/or
Arabic. "Verified purchase" is decided by the server: true once an order with
that product was delivered. Helpful votes: once per review, not on your own.
Sorting: newest, helpful, rating high/low.

### 6.12 Favorites and addresses
- Favorites: adding twice is harmless (201 new / 200 already saved); each
  favorite embeds the product with its live price.
- Addresses: the first becomes default; exactly one default (unique index +
  user-row lock); deleting the default promotes the newest remaining one.

### 6.13 Waiting list
- Customers can join only for out-of-stock products (one active request per
  product, enforced with an advisory lock).
- Statuses: `waiting ⇄ notified → fulfilled / cancelled`.
- **Automatic email**: when an admin saves the product with stock above 0,
  every `waiting` request becomes `notified` and the customer gets a "back in
  stock" email. Each request is told once; cancelled ones never.
- An admin can also mark a single request `notified` by hand, which sends the
  email too.

### 6.14 Notification emails and the email queue

| Email | When | How it is sent |
|---|---|---|
| Confirm your email | sign-up, "send again" | directly, right after the response |
| Reset your password | forgot password | directly, right after the response |
| Order #N received | checkout | queue |
| Order #N: confirmed / processing / shipped / delivered / cancelled | each status change, customer cancel | queue |
| Payment received | admin records a payment | queue |
| Account balance updated | admin records an adjustment | queue |
| Friendly reminder: open balance | periodically while a balance is open | queue |
| X is back in stock | product restocked / request marked notified | queue |

**Why a queue (the `email_outbox` table)?** If the mail provider is slow or
down, a checkout must not hang or fail. So services only *store* the email
(one fast INSERT), and a background worker delivers it:

```
service ──INSERT──► email_outbox (sent_at NULL, attempts 0, next_attempt_at now)
                          │
notify.Worker every EMAIL_POLL_INTERVAL (5 s):
   SELECT due rows ... FOR UPDATE SKIP LOCKED LIMIT 20
   send each one
     ok    → sent_at = now
     error → attempts + 1, last_error, next_attempt_at = now + 1, 2, 4, 8 ... min (max 6 h)
   after 10 failed attempts the row stays for an admin to look at
```

- Emails survive restarts (they are rows, not memory).
- `FOR UPDATE SKIP LOCKED` lets several API instances share the queue without
  sending anything twice.
- Why are confirmation and reset emails *not* queued? The queue stores the
  email text, and their text contains the secret link. Storing it would undo
  "only a hash of the token is stored".
- **Debt reminders**: about every 10 minutes the worker looks for active
  customers with a balance > 0 whose last reminder - or, the first time, last
  charge - is older than `DEBT_REMINDER_INTERVAL` (default 168 h = weekly).
  The reminder and `users.debt_reminded_at` are written in one transaction,
  so nobody is reminded twice for the same period. Paid-up customers get
  nothing. `DEBT_REMINDER_INTERVAL=0` turns reminders off.
- Useful SQL for an admin:
  ```sql
  SELECT kind, to_address, attempts, last_error, created_at FROM email_outbox WHERE sent_at IS NULL;  -- stuck emails
  UPDATE email_outbox SET attempts = 0, next_attempt_at = NOW() WHERE sent_at IS NULL;            -- retry now
  ```

## 7. Database and migrations

| Migration | Adds |
|---|---|
| `000001_baseline` | the original schema: users, products, product_images, product_colors, carts, cart_items, favorites, addresses, orders, order_items, reviews, review_helpful_votes, waiting_list_items |
| `000002_offers_bundles_debts` | offers, offer_products, bundles, bundle_items, cart_bundles, order_bundles, order_bundle_items, users.credit_limit, customer_debt_transactions, `pay_later` payment method, `deferred` payment status |
| `000003_auth_security` | users.token_version, password_reset_tokens |
| `000004_email_verification_notifications` | users.email_verified_at (existing users set to verified), users.debt_reminded_at, email_verification_tokens, email_outbox |

```powershell
go run ./cmd/migrate status   # read-only
go run ./cmd/migrate up       # apply what is pending - run after every pull that adds a migration
```

Each migration has a `.down.sql` to undo it by hand if ever needed. With
Docker Compose, migrations run automatically before the API starts.

How the code respects the schema: validators check `varchar` lengths first
(clear 400 instead of a database error); decimals match `numeric(10,2)`;
unique indexes (one default address, one primary image, one review per
product, one vote per review) get friendly 409 messages; `ON DELETE RESTRICT`
on orders is why accounts are deactivated rather than deleted.

## 8. Configuration and production mode

All settings are environment variables (or `.env` locally). Full lists with
comments: [.env.example](../.env.example) (development) and
[.env.production.example](../.env.production.example).

| Group | Variables |
|---|---|
| Basics | `APP_ENV` (development / production), `APP_PORT`, `LOG_FORMAT` (text / json; json is the production default) |
| Database | `DATABASE_URL`, `DB_MAX_CONNS` |
| Logins | `JWT_SECRET` (≥ 32 chars), `JWT_TTL` |
| Browser access | `CORS_ALLOWED_ORIGINS`, `TRUST_PROXY` |
| HTTPS | `TLS_MODE`, `TLS_DOMAINS`, `TLS_EMAIL`, `TLS_CACHE_DIR`, `TLS_CERT_FILE`, `TLS_KEY_FILE`, `TLS_PORT`, `TLS_HTTP_PORT` |
| Limits | `RATE_LIMIT_ENABLED` (the numbers are in `config.DefaultRateLimits`) |
| Shop | `STORE_NAME`, `FRONTEND_URL`, `CURRENCY` |
| Email | `MAIL_DRIVER`, `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM` |
| Links & tokens | `PASSWORD_RESET_URL`, `PASSWORD_RESET_TTL`, `EMAIL_VERIFY_URL`, `EMAIL_VERIFY_TTL`, `REQUIRE_EMAIL_VERIFICATION` |
| Queue | `EMAIL_POLL_INTERVAL`, `DEBT_REMINDER_INTERVAL` |
| Uploads | `PUBLIC_URL`, `STORAGE_DRIVER`, `UPLOAD_DIR`, `MAX_UPLOAD_MB`, `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_PUBLIC_URL`, `S3_INSECURE` |

**Production mode** (`APP_ENV=production`) refuses to start - and prints every
problem at once - when:

| Problem | Why it matters |
|---|---|
| `JWT_SECRET` is the example value or too predictable | anyone who knows it can sign in as any user, including admins |
| `CORS_ALLOWED_ORIGINS` is `*`, not https, or localhost | other websites could use visitors' logins; localhost cannot be your real shop |
| `FRONTEND_URL`, `PASSWORD_RESET_URL`, `EMAIL_VERIFY_URL` not public https | emailed links carry secrets |
| `PUBLIC_URL` / `S3_PUBLIC_URL` not public https | browsers would refuse or warn on images |
| `MAIL_DRIVER` is not smtp | the log driver writes reset links into the logs, and nobody receives mail |
| `RATE_LIMIT_ENABLED=false` | password guessing becomes unlimited |
| `S3_INSECURE=true` | uploads would travel unencrypted |

And it **warns** (logs, keeps running) about: a database connection with
`sslmode=disable` to a remote server, the `postgres` superuser in
`DATABASE_URL`, `JWT_TTL` over 7 days, links whose domain is not in the CORS
list, `TLS_MODE=off` with `TRUST_PROXY=false` (probably missing proxy
setting), and `TLS_MODE` on together with `TRUST_PROXY=true`.

## 9. HTTPS

Pick **one**:

| Option | Use when | Settings |
|---|---|---|
| **A. Caddy** (in `docker-compose.yml`) | your own server (VPS) - recommended | `API_DOMAIN` in `.env.production`; Compose sets `TLS_MODE=off`, `TRUST_PROXY=true` |
| **B. Hosting platform** | Render, Railway, Fly.io, Heroku-like hosts give HTTPS automatically | `TLS_MODE=off`, `TRUST_PROXY=true` |
| **C. The API itself** | a server without a proxy | `TLS_MODE=auto`, `TLS_DOMAINS=api.example.com`, ports 80 and 443 open; or `TLS_MODE=files` with your certificate |

Option C in detail (`internal/server`):
- `auto`: Let's Encrypt certificates are requested on first visit and renewed
  automatically; only names in `TLS_DOMAINS` are ever requested; certificates
  are cached in `TLS_CACHE_DIR` so restarts do not hit Let's Encrypt's limits.
- Port 80 answers Let's Encrypt's checks and redirects everything else to
  HTTPS with **308** (keeps POST as POST).
- TLS 1.2 is the minimum.

In every option, production responses carry **HSTS**
(`Strict-Transport-Security: max-age=63072000`) on HTTPS: after one visit,
browsers refuse plain HTTP to the API for two years, so nobody on the network
can downgrade the connection. The DNS record (A/AAAA) of the domain must
point at the server for options A and C.

## 10. Packaging and deployment

**The image** (`Dockerfile`): two-stage build. Stage 1 compiles static
binaries (`api`, `migrate`) with Go. Stage 2 is
`distroless/static:nonroot`: no shell, no package manager, runs as an
unprivileged user, contains only the binaries, the SQL migrations, CA
certificates and two writable folders (`/app/data/uploads`,
`/app/data/certs`). A `HEALTHCHECK` calls `api healthcheck`.

**One server with Compose** (`docker-compose.yml`):

```
db (postgres:18, volume db-data)  ──healthy──►  migrate (runs "migrate up", exits)
                                                    │ completed
                                                    ▼
                         caddy :80/:443  ──►  api :8080 (volume uploads)
```

Only Caddy is reachable from outside. Steps (full checklist in
[production.md](production.md)):

```bash
cp .env.production.example .env.production   # fill in every <...>
docker compose --env-file .env.production up -d --build
docker compose logs -f api                    # "starting env=production"
```

Updating: `git pull` then the same `up -d --build` (migrations run first,
then the new API replaces the old one). Back up the `db-data` volume (or use
`pg_dump`) daily, and keep `caddy-data` (certificates) and `uploads`.

Without Docker: `go build ./cmd/api ./cmd/migrate`, run `migrate up`, then
`api` under systemd or similar, behind Caddy or Nginx.

## 11. Tests

```powershell
go test ./...                                         # unit tests, no database

$env:INTEGRATION_DATABASE_URL = "postgres://USER:PASS@localhost:5432/postgres?sslmode=disable"
go test ./tests -v                                    # end-to-end on a throwaway database
```

The end-to-end tests create `ecommerce_it_<timestamp>`, apply every
migration, run the real router, and drop the database at the end - your real
data is never touched. Emails go to an in-memory mailer.

| Unit tests | Cover |
|---|---|
| `config` | defaults, every production refusal and warning, origin normalising, TLS settings |
| `images` | GPS removed, orientation kept, PNG/WebP metadata removed, fake/huge/broken images refused |
| `notify` | retry delays, email contents, link building |
| `server` | HTTP → HTTPS redirect, security headers, HSTS only on HTTPS in production |
| `auth`, `middleware`, `httpx`, `ratelimit`, `pricing`, `validators`, `models` | tokens, revocation, client IP, buckets, money math, input rules, status rules |

| End-to-end test | Checks |
|---|---|
| `TestFullShoppingFlow` | register → catalog → cart → checkout → statuses → reviews → waiting list, permissions |
| `TestLastUnitCanOnlyBeSoldOnce` | two simultaneous checkouts for the last unit |
| `TestAccountSecurity`, `TestLoginRevocation`, `TestPasswordReset`, `TestRateLimits` | account protection |
| `TestOffers`, `TestBundles`, `TestCustomerDebts` | promotions and pay later |
| `TestEmailVerification` | blocked checkout, resend, old/expired/used links, then checkout works |
| `TestOrderNotifications` | received / confirmed / shipped / cancelled emails, queue emptied |
| `TestEmailDeliveryIsRetried` | simulated SMTP outage: failure recorded, retried later, delivered once |
| `TestDebtNotificationsAndReminders` | payment and adjustment emails, reminder timing, never twice, none when paid |
| `TestBackInStockEmails` | restock emails waiting customers once, not cancelled ones |
| `TestImageUploads` | upload, serve, refusals (fake, too big, missing, not admin), deletes remove files, profile pictures |
| `TestMigrateAdoptsExistingDatabase`, `TestCORSPreflightAndHealth` | tooling |

## 12. API reference

Access: **P** public · **S** public with the strict account rate limit · **U**
logged in · **V** logged in with a confirmed email · **A** admin.

**Health & account**

| Method | Path | Access | Body / notes |
|---|---|---|---|
| GET | `/`, `/api/health` | P | health also pings the database |
| POST | `/api/register` | S | `first_name, last_name, email, password (8-72), phone, profile_image_url?` → 201 user; sends confirmation email |
| POST | `/api/login` | S | `email, password` → `{user, token, expires_at}` |
| POST | `/api/email/verify` | S | `token` (from the emailed link) |
| POST | `/api/email/verify/resend` | U | sends a new link (3 per hour) |
| POST | `/api/password/forgot` | S | `email` → always the same message |
| POST | `/api/password/reset` | S | `token, new_password` → all sessions revoked |
| POST | `/api/logout-all` | U | revoke every token |
| GET / PUT | `/api/profile` | U | PUT: `first_name, last_name, phone, profile_image_url?`; profile has `email_verified` |
| PUT | `/api/profile/password` | U | `current_password, new_password` → `{message, token, expires_at}` |
| POST | `/api/profile/image` | U | multipart, field `image` → updated user |
| DELETE | `/api/profile/image` | U | removes the picture |
| DELETE | `/api/profile` | U | deactivate |

**Catalog**

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/products` | P | `?category= &search= &sort= &in_stock= &on_offer= &limit= &offset=` |
| GET | `/api/products/{id}` | P | with images and colors |
| GET | `/api/products/{id}/images`, `/colors`, `/reviews` | P | reviews: `?sort=newest\|helpful\|rating_high\|rating_low` |
| GET | `/api/categories` | P | with counts |
| GET | `/uploads/{name}` | P | uploaded images (local storage) |
| GET | `/api/admin/products`, `/api/admin/products/{id}` | A | includes inactive |
| POST / PUT / DELETE | `/api/products`, `/api/products/{id}` | A | PUT with stock > 0 emails the waiting list |
| POST | `/api/products/{id}/images` | A | by URL: `image_url, is_primary, display_order` |
| POST | `/api/admin/products/{id}/images/upload` | A | multipart: `image` file, `is_primary?`, `display_order?` |
| PUT | `/api/product-images/{id}/primary` | A | |
| DELETE | `/api/product-images/{id}` | A | also deletes an uploaded file |
| POST | `/api/products/{id}/colors` | A | `name, name_ar, hex_code` |
| DELETE | `/api/product-colors/{id}` | A | |

**Shopping**

| Method | Path | Access | Notes |
|---|---|---|---|
| GET / DELETE | `/api/cart` | U | DELETE empties it |
| POST | `/api/cart/items` | U | `product_id, quantity?, selected_color?` |
| PUT / DELETE | `/api/cart/items/{id}` | U | |
| POST | `/api/cart/bundles` | U | `bundle_id, quantity?` |
| PUT / DELETE | `/api/cart/bundles/{id}` | U | |
| GET / POST | `/api/favorites` | U | POST `product_id` |
| DELETE | `/api/favorites/{product_id}` | U | |
| GET / POST | `/api/addresses` | U | |
| GET / PUT / DELETE | `/api/addresses/{id}` | U | |
| PUT | `/api/addresses/{id}/default` | U | |

**Orders**

| Method | Path | Access | Notes |
|---|---|---|---|
| POST | `/api/orders` | **V** | `address_id?, payment_method? (cash_on_delivery \| pay_later)`; 403 until the email is confirmed |
| GET | `/api/orders`, `/api/orders/{id}` | U | |
| POST | `/api/orders/{id}/cancel` | U | while pending |
| GET | `/api/admin/orders`, `/api/admin/orders/{id}` | A | `?status= &payment_method=` |
| PUT | `/api/admin/orders/{id}/status` | A | `status, payment_collected?` |

**Reviews, waiting list, offers, bundles, debts**

| Method | Path | Access |
|---|---|---|
| POST | `/api/products/{id}/reviews` | U |
| PUT / DELETE | `/api/reviews/{id}` | author |
| POST / DELETE | `/api/reviews/{id}/helpful` | U |
| POST / GET | `/api/waiting-list` | U |
| DELETE | `/api/waiting-list/{id}` | U |
| GET | `/api/admin/waiting-list` | A |
| PUT | `/api/admin/waiting-list/{id}/status` | A |
| GET | `/api/offers`, `/api/offers/{id}` | P |
| GET / POST | `/api/admin/offers` | A |
| GET / PUT / DELETE | `/api/admin/offers/{id}` | A |
| GET | `/api/bundles`, `/api/bundles/{id}` | P |
| GET / POST | `/api/admin/bundles` | A |
| GET / PUT / DELETE | `/api/admin/bundles/{id}` | A |
| GET | `/api/debt` | U |
| GET | `/api/admin/debts` | A |
| GET | `/api/admin/customers/{id}/debt` | A |
| PUT | `/api/admin/customers/{id}/credit-limit` | A |
| POST | `/api/admin/customers/{id}/debt/payments` | A |
| POST | `/api/admin/customers/{id}/debt/adjustments` | A |

Bodies for offers, bundles and debts: [offers-bundles-debts.md](offers-bundles-debts.md).

**Status codes**: 400 bad input (see `fields`), 401 not logged in / expired /
revoked / deactivated, 403 not allowed (not admin, or email not confirmed for
checkout), 404 not found or not yours, 409 not possible in the current state,
429 too many requests (`Retry-After`), 500 unexpected (details only in the
log). Lists: `?limit=` (1-100, default 20) and `?offset=` →
`{items, total, limit, offset}`. Money: JSON numbers with ≤ 2 decimals.

## 13. What the frontend must do

- Send `Authorization: Bearer <token>`. On 401, delete the token and show the
  login page (also for "this login has been revoked").
- After changing the password, store the new `token` from the response.
- Show 429 messages; wait `Retry-After` seconds before retrying.
- **Pages the emails link to** (relative to `FRONTEND_URL`):
  - `/verify-email?token=...` → `POST /api/email/verify {token}` → "email confirmed"
  - `/reset-password?token=...` → new-password form → `POST /api/password/reset`
  - `/orders/{id}` - order detail
  - `/account/debt` - balance and statement (`GET /api/debt`)
  - `/products/{id}` - product page
- Show a banner while `profile.email_verified` is false, with a "send again"
  button (`POST /api/email/verify/resend`). On 403 at checkout, show the
  message from the API.
- Uploads: send `FormData` with the file in field `image` (do **not** set
  `Content-Type` yourself; the browser adds the boundary). Accept
  `image/jpeg, image/png, image/webp`, max 5 MB by default.

## 14. Limits and next steps

- Rate-limit counters are in memory: with several API instances, move them to
  Redis (only `internal/ratelimit` changes).
- Uploaded images are stored as sent (after cleaning); there is no automatic
  resizing or thumbnail generation - a CDN with image resizing (Cloudflare
  Images, imgproxy) is the usual next step.
- Online card payments, coupon codes typed at checkout, SMS notifications and
  a CAPTCHA are not included.
- Emails that failed 10 times stay in `email_outbox` for an admin to inspect;
  there is no admin screen for them yet.
