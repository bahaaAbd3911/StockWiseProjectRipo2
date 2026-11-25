# StockWise — the complete project guide

This document explains the whole project: what it does, how every part
works, and **why** it was built this way. It is written to be read from top to
bottom once; after that, use the contents to jump to a part.

Deeper, more technical references:

| Document | What is in it |
|---|---|
| [backend/docs/GUIDE.md](backend/docs/GUIDE.md) | backend reference: every package, every endpoint with its body, all settings |
| [backend/docs/production.md](backend/docs/production.md) | step-by-step server checklist |
| [backend/docs/offers-bundles-debts.md](backend/docs/offers-bundles-debts.md) | offers, bundles and pay-later in detail |
| [backend/docs/auth-security.md](backend/docs/auth-security.md) | rate limits, login revocation, password reset in detail |

## Contents

1. [What StockWise is](#1-what-stockwise-is)
2. [The big picture](#2-the-big-picture)
3. [What is in the repository](#3-what-is-in-the-repository)
4. [How the project got here](#4-how-the-project-got-here)
5. [One request, from click to database](#5-one-request-from-click-to-database)
6. [Many stores in one system (multi-tenancy)](#6-many-stores-in-one-system-multi-tenancy)
7. [The backend](#7-the-backend)
8. [Every feature and its logic](#8-every-feature-and-its-logic)
9. [The frontend](#9-the-frontend)
10. [The database and its migrations](#10-the-database-and-its-migrations)
11. [Settings](#11-settings)
12. [Running it on your computer](#12-running-it-on-your-computer)
13. [Running it on a server](#13-running-it-on-a-server)
14. [Tests](#14-tests)
15. [Security at a glance](#15-security-at-a-glance)
16. [Decisions and why](#16-decisions-and-why)
17. [Limits and what could come next](#17-limits-and-what-could-come-next)
18. [Words used in this document](#18-words-used-in-this-document)

---

## 1. What StockWise is

StockWise is an **online-shop platform**. One codebase and one deployment run
**many independent stores**. Each store has its own address, products,
customers, orders, prices, coupons, logo, colors and page content - and none
of them can see or touch another store's data.

There are three kinds of people:

| Who | Where | What they do |
|---|---|---|
| **Customers** | the store's address, e.g. `ahmad.myplatform.com` or `ahmadstore.com` | browse, buy (cash on delivery or pay later), track orders, review, keep favorites |
| **Store admins** | the same address, admin pages (`/admin`) | manage products, categories, orders, customers, coupons, offers, debts, the look of the store, their domains |
| **Platform admins** | `admin.myplatform.com` | create stores (with their first admin), edit, suspend and reactivate them |

Everything is bilingual (English and Arabic, with right-to-left layout) and
has a light and a dark theme.

A new store is **ready the moment it is created**: no new database, no new
server, no new deployment. The platform admin fills in a form; the store's
admin receives an email, chooses a password and starts selling.

## 2. The big picture

```
                 ahmad.myplatform.com   omar.myplatform.com   ahmadstore.com   admin.myplatform.com
                          \                    |                    |                /
                           \                   |                    |               /
                            v                  v                    v              v
                 ┌────────────────────────────────────────────────────────────────────┐
                 │  Caddy (web server)                                                │
                 │   - HTTPS certificates for every host (only hosts the API confirms)│
                 │   - serves the React app (the same files for every store)          │
                 │   - sends /api/* and /uploads/* to the Go API, keeping the host    │
                 └───────────────────────────────┬────────────────────────────────────┘
                                                 │
                 ┌───────────────────────────────v────────────────────────────────────┐
                 │  Go API                                                            │
                 │   1. which store?  from the host name                              │
                 │   2. who is it?    login token (must belong to that store)         │
                 │   3. what is allowed? customer / store admin / platform admin      │
                 │   4. business rules, then SQL                                      │
                 │   + background worker: sends queued emails, debt reminders         │
                 └───────────────────────────────┬────────────────────────────────────┘
                                                 │  as role stockwise_app, with app.store_id set
                 ┌───────────────────────────────v────────────────────────────────────┐
                 │  PostgreSQL (one database for all stores)                          │
                 │   every row has store_id; row-level security shows a connection   │
                 │   only the rows of its current store                               │
                 └────────────────────────────────────────────────────────────────────┘
```

The React app does not know any store id. It just calls `/api/...` on the
address it was opened from; the backend recognises the store from that
address. This is why one build of the frontend can serve every store and
every custom domain.

## 3. What is in the repository

```
Project/
├── Front-End/                  the React app (Vite)
│   ├── src/
│   │   ├── api/                every call to the backend, and the field conversions
│   │   ├── context/            app-wide state: session, store settings, catalog, cart
│   │   ├── hooks/              small reusable pieces of logic
│   │   ├── layouts/            the shop layout and the admin layout
│   │   ├── Pages/              screens: client/, admin/, account/, platform/
│   │   ├── components/         building blocks: client/, admin/ (UI kit), common/
│   │   ├── data/ constants/    default texts, settings defaults, labels
│   │   └── utils/              formatting and validation helpers
│   └── vite.config.js          development server; passes /api to the backend
├── backend/                    the Go API
│   ├── cmd/                    programs: api (the server), migrate, platform (create a platform admin)
│   ├── internal/               all the application code (see section 7)
│   ├── migrations/             the database changes, numbered 000001 ... 000009
│   ├── tests/                  end-to-end tests against a throwaway database
│   ├── deploy/                 Caddyfile, web-server image, backup and restore scripts
│   ├── docs/                   the backend references
│   ├── Dockerfile              the API image
│   └── docker-compose.yml      the whole platform on one server
├── package.json                shortcuts: npm run dev / api / build / lint
└── DOCUMENTATION.md            this file
```

## 4. How the project got here

| Step | What happened | Why |
|---|---|---|
| Original code | a first backend (products, cart, orders, reviews) and a React template working on sample data | the starting point |
| **Backend rewrite** | rebuilt against the real database schema in clear layers, with exact money, typed errors, validation, locks at checkout, a migration tool and tests | the first version mixed HTTP, rules and SQL, used floating-point money, and had race conditions (two people could buy the last item) |
| **Offers, bundles, pay later** | timed promotions with one best price, bundles sold as one item, credit limits with a debt ledger | requested features; designed so prices never disagree between pages |
| **Account security** | rate limits, revocable logins, password reset by email | needed before anything goes public |
| **Production readiness** | production safety checks, HTTPS, security headers, Docker, email confirmation, a durable email queue, image uploads, coupons, a dashboard, backups | what a real shop needs to run safely |
| **Frontend restructure** (teammate) | the React app was reorganised and connected to an API | the template used fake data |
| **Connecting the two** | the frontend's API layer was aligned with the backend; categories, store settings and admin user management were added to the backend | the two halves were written against different assumptions and did not fit |
| **Multi-store, step 1** | stores, host-name detection, isolation enforced by the database | one deployment, many independent stores |
| **Step 2** | platform super admin, store creation with an invite, per-store emails | stores must be created without a developer |
| **Step 3** | store customization: banners, feature tiles, footer, images | each store should look like its own brand |
| **Step 4** | custom domains with DNS verification, certificates on demand, per-store upload folders | stores want their own domain name |

## 5. One request, from click to database

Example: a customer of Ahmad's store adds a product to the cart.

1. **Browser** → `POST https://ahmad.myplatform.com/api/cart/items` with the
   login token in `Authorization: Bearer ...` and `{product_id, quantity}`.
2. **Caddy** has a certificate for `ahmad.myplatform.com` and forwards the
   request to the API, keeping the host name and adding the visitor's IP.
3. **API, global steps** (`internal/middleware`):
   - recover from crashes, log the request, add security headers;
   - find the visitor's IP (only trusting the proxy's header);
   - **find the store** from the host name: `ahmad` → store #12 (unknown → 404,
     suspended → 403);
   - general rate limit per IP.
4. **Route** `POST /api/cart/items` requires a logged-in user:
   - the token is checked (signature, expiry), **its store must be #12**, the
     user must exist in store #12, be active, and the token must not be revoked.
5. **Handler** reads the JSON → **service** validates (quantity 1-1000, color
   required if the product has colors) → **repository** runs the SQL.
6. **Database connection**: before the SQL, the connection is told
   `app.store_id = 12`. The database's policies then only let the query see
   and change rows of store #12. If the product id belongs to another store,
   the product simply is not found → 404.
7. The updated cart is returned as JSON, with live prices.

Errors travel back the same way: every layer returns a typed error ("not
found", "conflict", "invalid"...) and one function turns it into the HTTP
status and a JSON body `{"error": "...", "fields": {...}}`.

## 6. Many stores in one system (multi-tenancy)

### 6.1 The choice: one database, shared tables, `store_id` on every row

| Option | Isolation | Cost to run | Adding a store | Chosen? |
|---|---|---|---|---|
| A database per store | strongest | high: N databases to migrate, back up, connect to | create a database | no |
| A schema per store | strong | medium: N schemas to migrate | create a schema | no |
| **Shared tables + `store_id` + row-level security** | strong, enforced by PostgreSQL | **low: one database, one migration, one backup** | **insert one row** | **yes** |

The requirement was "one codebase, one deployment, a store ready without a new
database". Shared tables meet it directly. The usual weakness of shared tables
- a forgotten `WHERE store_id = ...` leaks data - is removed by making the
**database** enforce the filter (6.3).

### 6.2 Finding the store from the host name

`internal/tenant` (the resolver) and `middleware.Store`:

| Host | Store |
|---|---|
| `<slug>.PLATFORM_DOMAIN` | the store with that slug |
| a verified custom domain | the store that owns it |
| anything else (localhost, an IP, the bare domain) | `DEFAULT_STORE` (`main`), or 404 if empty |
| `PLATFORM_ADMIN_HOST` | no store: only the platform API |

Why the host and not a header or a path like `/store/ahmad`? Because the host
is what the visitor typed, the browser keeps logins separate per host, and it
works the same for subdomains and custom domains. **The store is never taken
from anything the client sends** (no `store_id` in bodies, headers or query
strings), so it cannot be forged. Answers are cached 30 seconds; suspending a
store clears the cache.

### 6.3 Three layers of isolation

1. **Row-level security (RLS).** The API connects and then switches to the
   database role `stockwise_app`, which is not the owner of the tables, so
   PostgreSQL's policies apply to it. Every store table has the policy
   "rows are visible and writable only if `store_id = app.store_id`". A
   connection with no store set sees **nothing** - a bug fails closed, it
   never leaks.
2. **`store_id` fills itself in.** Its default is the current store, so the
   existing `INSERT` statements needed no change, and inserting a row for
   another store is refused by the policy.
3. **Store-aware foreign keys.** Links include the store:
   `(store_id, product_id) → products (store_id, id)`. Even where policies do
   not apply (foreign-key checks ignore them), a cart line can never point at
   another store's product.

How the connection knows the store: `database.ConnectApp` wraps every
connection. Before each statement it compares the request's store with the
connection's current `app.store_id` and only sends a change when needed (so
it costs almost nothing). A transaction belongs to one store; switching store
inside one is an error.

**Platform mode** (`tenant.AsPlatform`) lets trusted server code see every
store: the email worker, debt reminders, clean-up jobs, and the platform
admin's API. It can only be switched on by code, never by request data.

### 6.4 Logins per store

- A login token records its store (`sid`). It is refused on any other store's
  host - "this login belongs to another store".
- Customer accounts are per store: the same email can have an account in two
  stores, with different passwords.
- Platform admins are a separate table with their own token issuer: a store
  admin's token is useless on the platform API and the other way round.
- Rate-limit counters (failed logins, reset emails) are per store too.

### 6.5 What is per store

Products, categories, offers, bundles, coupons, customers, carts, favorites,
addresses, orders, reviews, waiting lists, debts, emails in the queue, the
dashboard numbers, store settings and page content, uploaded files, custom
domains, the email sender name and links, currency.

### 6.6 Custom domains and HTTPS

1. The store admin adds `shop.example.com` → the API gives a random token.
2. At their DNS provider they add:
   - `TXT  _stockwise-verify.shop.example.com  "stockwise-verify=<token>"` (proof of ownership)
   - `CNAME shop.example.com → ahmad.myplatform.com` (sends visitors to the platform)
3. They press **Verify**; the API looks the TXT record up. Only a verified
   domain shows the store.
4. On the first visit, Caddy asks the API `/internal/tls-check?domain=...`
   whether to get a certificate. Only real store hosts (subdomains, verified
   domains of active stores, the platform admin, the bare domain) pass, so
   nobody can make the server request certificates for random names.
5. One verified domain can be the store's **main address**: emails then link to it.

Why a TXT record? Pointing a domain at the server proves nothing (anyone can
point a domain anywhere); publishing a secret token in that domain's DNS can
only be done by its owner.

## 7. The backend

### 7.1 Layers

```
middleware  ->  routes  ->  handlers  ->  services  ->  repository  ->  PostgreSQL
(every         (who may    (HTTP in/out)  (business     (SQL,
 request)       call what)                 rules)        transactions)
```

| Layer | Knows about | Never contains |
|---|---|---|
| handlers | HTTP, JSON | SQL, business rules |
| services | rules, validation, which repositories to combine | HTTP status codes, SQL |
| repository | SQL | HTTP |

Why layers? Each file has one job, so a change in the shop's rules touches a
service, a change in a query touches a repository, and nothing else moves.
It also makes each part testable on its own.

### 7.2 The packages (`backend/internal`)

| Package | Job |
|---|---|
| `tenant` | the store of a request; host name → store |
| `database` | connection pool, the store-aware connection wrapper, transactions |
| `config` | every setting; refuses unsafe production settings |
| `apperr`, `httpx` | typed errors; JSON in/out; error → status code |
| `auth` | login tokens (store and platform), the caller's identity |
| `middleware` | store detection, login checks, admin / platform checks, CORS, rate limits, security headers, logging, crash recovery |
| `routes` | builds everything and lists every endpoint |
| `handlers` | one file per area (products, cart, orders, coupons, domains, platform...) |
| `services` | the rules per area |
| `repository` | the SQL per area |
| `models` | data shapes and simple rules (which order status may follow which) |
| `validators` | cleans input (trims, lower-cases emails) and checks it, reporting every problem at once |
| `pricing` | the only place prices and discounts are calculated |
| `notify` | email texts (English + Arabic, per store), the email queue and its worker |
| `mail` | sending: print to the log (development) or SMTP |
| `storage` | uploaded files: local disk or S3, one folder per store |
| `images` | checks uploads are real images; removes GPS and other metadata |
| `ratelimit` | counters that slow down abuse |
| `server` | HTTP / HTTPS listeners, graceful stop |
| `migrate` | applies the SQL migrations in order |

### 7.3 Ideas used everywhere

- **Exact money**: prices are `decimal` in Go and `numeric(10,2)` in the
  database, never floating point (where 0.1 + 0.2 = 0.30000000000000004).
- **All-or-nothing writes**: anything touching several rows runs in one
  transaction; if a step fails nothing is saved.
- **Row locks against races**: checkout locks the products it sells, so the
  last item can only be sold once (a test fires two checkouts at the same time).
- **Not yours = not found**: personal data is always filtered by user (and by
  store), so another user's order id returns 404.
- **Soft deletes**: products, bundles and accounts are deactivated, not
  deleted, because orders refer to them; orders keep **copies** of names,
  prices and the address, so history never changes.
- **One rule, one place**: the price (`pricing`), whether an item can be
  bought (`cartLine.problem`), whether an offer is live, which order status can
  follow which - each lives in one function used by every screen.
- **Secrets are hashed**: passwords with bcrypt; email links (reset,
  confirmation, store invite) are random tokens of which only a hash is stored.

## 8. Every feature and its logic

### Accounts
- **Sign up** creates a customer of *this* store and emails a confirmation
  link. Logging in works right away; **placing an order needs a confirmed
  email** (so nobody can sign up with someone else's address and receive
  order emails meant for them).
- **Login** answers the same way, at the same speed, for unknown emails and
  wrong passwords, so it cannot be used to find out who is a customer.
- **Revocation**: changing or resetting the password, or "log out everywhere",
  makes all older tokens invalid (a version number in the token).
- **Password reset**: always the same reply; a single-use link valid 30 minutes.
- **Rate limits**: general per IP; stricter on login, sign-up, password and
  coupon endpoints; failed logins and reset emails are also counted per
  email, per store.

### Catalog
Products in English and Arabic, with category, price, own discount, stock,
images (uploaded or by URL), colors, and computed fields: final price, in
stock, rating. **Categories** are records with an Arabic name, an icon and an
on/off switch; renaming one renames it on its products, deleting one hides
its products (kept for order history).

### Images and uploads
Uploads are checked by their bytes (a "picture" that is really HTML is
refused), limited in size and dimensions, and **stripped of metadata** - phone
photos contain the GPS position where they were taken, often the customer's
home. Each store's files live in its own folder (`uploads/<store id>/`), and a
store can only delete its own files.

### Prices, offers, coupons, bundles
- **One best price, never stacked**: a product's own discount or the best
  live **offer**, whichever is lower. Every page uses the same function, so
  the product page, the cart and checkout always agree.
- **Coupons** are typed at checkout and come **on top of** offers (decided
  with you: simple to explain, controlled by minimum order and maximum
  discount). Limits: dates, total uses, uses per customer (default once).
  Checkout re-checks the coupon with its row locked, so two customers cannot
  both take its last use.
- **Bundles**: several products for one price, sold as one line; their stock
  is counted together with the same products bought alone.

### Cart and checkout
The cart always shows live prices and whether each line can still be bought.
Checkout is one transaction: address, locks, stock check of everything,
prices, coupon, credit limit (pay later), order with copies of every name and
price, stock reduced, cart emptied.

### Orders
`pending → confirmed → processing → shipped → delivered`, or `cancelled`.
Cancelling gives back stock, reverses debt and frees the coupon use;
delivering marks reviews as verified purchases and settles the payment.

### Pay later and debts
An admin gives a customer a credit limit. Pay-later orders add a **charge**
to an append-only ledger (nothing is ever edited; corrections are new rows),
payments and adjustments are rows too, and the balance is the sum. Weekly
reminders are emailed while a balance is open.

### Reviews, favorites, addresses, waiting list
One review per product per customer; "verified purchase" is decided by the
server. One default address. Customers can wait for out-of-stock products and
are emailed automatically when the stock comes back.

### Emails and the email queue
Order, payment, reminder and back-in-stock emails are **queued in the
database** and sent by a background worker that retries failures with growing
delays. A broken mail provider therefore never slows or fails a checkout, and
queued emails survive restarts. Emails that contain a secret link (reset,
confirmation, invite) are sent directly instead, so the secret is never
stored. Every email carries its store's name, links and currency. Admins see
the queue (pending, failed, sent) and can retry.

### Admin dashboard
One call returns today, the chosen period and the period before it (for
"+12%" comparisons), sales per day (in the store's time zone), best sellers,
low stock, debts, waiting-list demand, coupon use and customers - all from one
database snapshot so the numbers agree.

### Admin user management
Admins list, search, create, edit and deactivate users. Safety rules: an
admin cannot remove their own rights, and the last admin can never be removed.

### Store customization (step 3)
Each store's settings (one row per store) hold its name, logo, theme colors
and **page content**: home-page banner, products-page banner, feature tiles,
footer tagline, copyright and social links - in English and Arabic. Banner
images are uploaded. Empty fields keep the template's original texts, so a
new store looks complete from the start. The same React components render
every store from these settings; no component changes per store.

### Platform super admin (step 2)
Creates a store in one step (store record, settings, admin account, invite
email), lists stores with their numbers, edits, suspends (everyone gets "this
store is currently unavailable", nothing is deleted) and reactivates. There is
no public sign-up for platform admins: the first one is created on the server
with `go run ./cmd/platform create-admin`.

## 9. The frontend

### 9.1 Structure

| Folder | Role |
|---|---|
| `api/` | the **only** place that talks to the backend. `client.js` sends requests (to `/api` on the same address), reads errors, fetches every page of long lists, uploads files. `mappers.js` converts the backend's snake_case fields into the shapes the components use. |
| `context/` | `AuthProvider` (session, language, theme, messages), `StoreProvider` (this store's settings: name, logo, colors, content), `CatalogProvider` (products and categories), `ShopProvider` (cart, favorites, orders) |
| `layouts/` | the shop layout (with the "confirm your email" banner) and the admin layout (sidebar, guard that only lets admins in) |
| `Pages/client` | home, products, product details, cart, checkout, profile, favorites, orders |
| `Pages/admin` | dashboard, products, categories, orders, users, store settings, domains |
| `Pages/account` | confirm email, reset password (opened from emails) |
| `Pages/platform` | the platform admin: login and stores (shown on `admin.*` hosts) |
| `components/admin/ui` | the admin UI kit: Button, Card, Modal, FormField, tables, badges... |

### 9.2 Why the API layer and mappers

The components were written with their own field names (`nameAr`, `count`,
`image`...). Instead of rewriting every component to the backend's names,
`mappers.js` translates in one place. When the backend changes a field, only
the mapper changes.

### 9.3 One build, every store

`main.jsx` checks the host: on `admin.*` it shows the platform section,
anywhere else the shop. The shop loads `/api/settings` (this store's settings)
and applies its colors (CSS variables), logo, name and content. Nothing in the
build is specific to a store.

In development the Vite server passes `/api` and `/uploads` to the backend
and keeps the host name, so `ahmad.localhost:5173` reaches store `ahmad`.

## 10. The database and its migrations

Migrations are numbered SQL files applied in order by `go run ./cmd/migrate up`
(recorded in `schema_migrations`, so each runs once). With Docker they run
automatically before the API starts. Each has a `down` file to undo it by hand.

| # | Adds | Why |
|---|---|---|
| 000001 baseline | the original schema: users, products, images, colors, carts, favorites, addresses, orders, reviews, waiting list | the starting point |
| 000002 | offers, bundles, coupons-to-be, credit limits, debt ledger, pay later | promotions and pay later |
| 000003 | token version, password reset tokens | revocable logins, password reset |
| 000004 | email confirmation, email queue, reminder dates | trusted emails, reliable notifications |
| 000005 | coupons, redemptions, order discount fields, dashboard indexes | coupon codes, fast dashboard |
| 000006 | categories, store settings | what the admin panel manages |
| 000007 | stores, store domains, `store_id` everywhere, store-aware foreign keys, per-store unique rules, row-level security, role `stockwise_app` | many stores in one database |
| 000008 | platform admins, sender name on queued emails | platform super admin, per-store emails |
| 000009 | domain verification token, one main domain per store | custom domains |

## 11. Settings

All settings are environment variables (or `backend/.env` locally). The full
commented lists are `backend/.env.example` (development) and
`backend/.env.production.example` (a server). The important groups:

| Group | Settings |
|---|---|
| Basics | `APP_ENV` (development / production), `APP_PORT`, `DATABASE_URL`, `JWT_SECRET` |
| Stores | `PLATFORM_DOMAIN`, `DEFAULT_STORE`, `PLATFORM_ADMIN_HOST`, `STORE_URL` |
| Addresses | `FRONTEND_URL`, `PUBLIC_URL`, `CORS_ALLOWED_ORIGINS`, `TRUST_PROXY` |
| Email | `MAIL_DRIVER` (log / smtp), `SMTP_*`, link lifetimes, reminders |
| Uploads | `STORAGE_DRIVER` (local / s3), `UPLOAD_DIR`, `S3_*`, `MAX_UPLOAD_MB` |
| Shop | `STORE_NAME` (fallback), `CURRENCY` (fallback), `STORE_TIMEZONE`, `LOW_STOCK_THRESHOLD` |

With `APP_ENV=production` the API **refuses to start** with unsafe settings
(an example secret, localhost addresses, emails printed to the log, rate limits
off...) and lists everything to fix at once.

## 12. Running it on your computer

Requirements: Go, Node.js, PostgreSQL.

```powershell
cd backend
copy .env.example .env        # set DATABASE_URL and a JWT_SECRET of 32+ characters
go run ./cmd/migrate up       # create / update the database
go run ./cmd/api              # the API on http://localhost:8080

cd ..\Front-End
npm install
npm run dev                   # the shop on http://localhost:5173
```

**One shop:** that is all - `localhost` is the `main` store. Emails are
printed in the API's terminal (copy confirmation and reset links from there).

**Several stores and the platform admin:** add `PLATFORM_DOMAIN=localhost` to
`backend/.env`, restart the API, then:

```powershell
go run ./cmd/platform create-admin -email you@example.com    # prints a password
```

Open `http://admin.localhost:5173`, sign in, create a store `ahmad`, and open
`http://ahmad.localhost:5173`. (Browsers send every `*.localhost` name to your
own computer.)

## 13. Running it on a server

`backend/docker-compose.yml` runs everything on one server:

```
db (PostgreSQL) → migrate (applies migrations, exits) → api → web (Caddy + the React app)
```

Only `web` is reachable from outside (ports 80 and 443). Steps (full list in
[production.md](backend/docs/production.md)):

1. DNS: `A` records for `myplatform.com` and `*.myplatform.com` → the server.
2. `cp .env.production.example .env.production` and fill it in (domain,
   database password, JWT secret, SMTP).
3. `docker compose --env-file .env.production up -d --build`
4. Create the first platform admin (prints a password once):
   `docker compose --env-file .env.production exec api /app/platform create-admin -email you@example.com`
5. Backups: `deploy/backup.sh` daily (database + uploads, copied off the
   server), and prove them with `deploy/restore.sh --test`.

Certificates are issued automatically per host (6.6). With many stores, use
one wildcard certificate for `*.myplatform.com` instead (production.md).

## 14. Tests

```powershell
cd backend
go test ./...                                          # unit tests
$env:INTEGRATION_DATABASE_URL = "postgres://USER:PASS@localhost:5432/postgres?sslmode=disable"
go test ./tests -v                                     # end-to-end
```

The end-to-end tests create a **throwaway database**, apply every migration,
run the real API (as the restricted database role, like production), and drop
the database at the end; your real data is never touched. They cover the
whole shopping flow, the last-item race, security, offers, bundles, coupons,
debts, emails, uploads, the dashboard, and the multi-store guarantees: one
store trying to read, change or link to another store's data through the API,
with edited ids, with another store's token, and even with raw SQL on the
API's own database connection.

Frontend: `npm run lint` and `npm run build` in `Front-End`.

## 15. Security at a glance

| Risk | Protection |
|---|---|
| one store reading another's data | store from the host only; row-level security; store-aware foreign keys; per-store tokens |
| guessing passwords | bcrypt; rate limits per IP and per email; same answer for unknown emails |
| stolen or old logins | short-lived tokens; revocation by version; tokens bound to their store |
| fake sign-ups receiving emails | email confirmation before ordering |
| leaked password-reset links | random tokens, only hashes stored, single use, short life |
| overselling / double use of a coupon | row locks in one transaction |
| malicious uploads, GPS in photos | content checks, size limits, metadata stripped, random names |
| scripts in store settings | values validated (image URLs, Font Awesome icons only, known social networks, hex colors) |
| plain-text traffic | HTTPS everywhere, HSTS, HTTP redirected |
| certificates for random domains | Caddy asks the API first; only verified, active store hosts |
| unsafe production config | the API refuses to start and lists the problems |

## 16. Decisions and why

| Decision | Alternatives | Why this one |
|---|---|---|
| Go backend in layers | keep the original code; Node.js | the existing backend was Go; layers make it testable and clear; rewriting in another language would have thrown work away |
| Money as decimals | floats | floats give rounding errors on prices |
| PostgreSQL row locks at checkout | check-then-update without locks | without locks two buyers can both get the last item |
| Best single discount, no stacking (offers) | add discounts up | predictable prices; protects margins |
| Coupons on top of offers | never with offers; best of the two | simplest to explain; the admin controls it with minimum order and caps (your choice) |
| Append-only debt ledger | editable balance column | a full history of every charge and payment; mistakes are corrected visibly |
| Email queue in the database | send during the request | a slow mail provider must not break checkout; emails survive restarts; no extra service (like Redis) needed |
| Secret-link emails sent directly | queue them too | the queue would store the secret link |
| Email confirmation required for checkout only | block login too | customers can browse and fill a cart right away; only ordering needs a real email |
| Frontend API layer with mappers | rewrite components to backend names | keeps the existing UI untouched; one place to adapt |
| One database, shared tables, RLS | database or schema per store | no per-store infrastructure; a store is one row; RLS makes forgotten filters harmless |
| Store from the host name | store id in the request | cannot be forged; works for subdomains and custom domains; browsers keep logins per host |
| API on the same address as the shop (`/api`) | a separate API domain | the host decides the store, no CORS, one frontend build for all stores |
| Separate platform admins table and tokens | a "super admin" role on users | a store login can never become a platform login |
| Invite email instead of a password | platform admin types a password | nobody but the store admin ever knows their password |
| TXT record for domain ownership | trust the CNAME | pointing a domain proves nothing; only the owner can publish the token |
| Certificates on demand with an "ask" check | wildcard only; manual certificates | works for custom domains with no setup; the check blocks abuse |
| Page content as data with defaults | per-store component changes | one codebase; a new store looks complete; admins change only what they want |
| Per-store upload folders + ownership check | one shared folder | files are manageable per store; no store can delete another's file |
| Microservices: not done | split into services | the platform works well as one service; splitting would add network failures, many deployments and break the one-transaction checkout without a real need (revisit if there is a requirement) |

## 17. Limits and what could come next

- **Before going live:** build and test the Docker images once (Docker
  Desktop did not run during development), run the backup scripts for real,
  set up email (SMTP with SPF/DKIM), monitoring and backups.
- **Many stores:** switch store subdomains to a wildcard certificate (Let's
  Encrypt allows 50 certificates per week per domain); move rate-limit
  counters to Redis when running several API servers.
- **Frontend screens still missing** for features the backend already has:
  coupon box in the cart, offers, bundles, pay-later debts and the waiting list
  for customers; coupons, offers, bundles, debts and the email queue for admins.
  Product pages show the product's own discount, not live offer prices (the
  cart and checkout already use the right price).
- **Later:** online card payments, SMS, image resizing / CDN, CAPTCHA.

## 18. Words used in this document

| Word | Meaning |
|---|---|
| tenant / store | one independent shop on the platform |
| slug | a store's short name in its address (`ahmad` in `ahmad.myplatform.com`) |
| platform domain | the domain the platform runs on (`myplatform.com`) |
| custom domain | a store's own domain name (`ahmadstore.com`) |
| row-level security (RLS) | a PostgreSQL feature that filters rows per connection, below the application |
| migration | a numbered SQL file that changes the database structure |
| token (JWT) | the signed proof of login the browser sends with each request |
| transaction | several database changes that succeed or fail together |
| row lock | makes other requests wait for a row until the current change is finished |
| on-demand TLS | getting an HTTPS certificate the first time a host is visited |
| CNAME / TXT | DNS record types: an alias to another name / a piece of text |
