# 2. How the project was built

The project grew in steps. Each step solved a concrete problem of the step
before it. This file tells the story in order, with the reason for each step.

## Step 0: the starting point

- **Backend:** a first Go API with products, cart, orders, order statuses and
  reviews.
- **Frontend:** a React shop template that worked on sample data written into
  the code.

Problems found in the original backend:

- HTTP handling, business rules and SQL were mixed in the same functions, so
  a change in one place broke others and nothing could be tested alone.
- Money used floating-point numbers, which produce rounding errors
  (0.1 + 0.2 = 0.30000000000000004).
- Checkout checked stock and then updated it without locking. Two customers
  could both buy the last item.
- Errors were plain strings, and database messages could reach the browser.

## Step 1: the backend rewrite

The backend was rebuilt against the real database schema:

- **Layers**: handlers (HTTP) → services (rules) → repository (SQL). See
  [06-backend-architecture.md](06-backend-architecture.md).
- **Exact money**: `decimal` in Go, `numeric(10,2)` in the database.
- **Typed errors** that decide the HTTP status. Unknown errors show only
  "internal server error"; the details go to the log.
- **Validation** that reports every bad field at once.
- **Row locks at checkout**, so the last item is sold exactly once. A test
  fires two checkouts at the same moment to prove it.
- **A migration tool** (`go run ./cmd/migrate up`), so the database structure
  is versioned with the code.
- **Unit tests and end-to-end tests** on a throwaway database.

## Step 2: offers, bundles and pay later (migration 000002)

The shop needed promotions and credit for trusted customers:

- **Offers**: timed price cuts on products or a whole category.
- **Bundles**: several products sold together for one price.
- **Pay later**: credit limits and an append-only debt ledger.

The key rule decided here: a customer always gets **one best price** and
discounts never add up, and one function computes it for every page. See
[09-prices-offers-coupons-bundles.md](09-prices-offers-coupons-bundles.md)
and [11-pay-later-and-debts.md](11-pay-later-and-debts.md).

## Step 3: account security (migration 000003)

Before anything goes public:

- rate limits (per IP and per email);
- logins that can be revoked ("log out everywhere", and automatically after
  a password change);
- password reset by email.

See [07-accounts-and-security.md](07-accounts-and-security.md).

## Step 4: production readiness (migrations 000004 and 000005)

What a real shop needs to run safely:

- a **production mode** that refuses to start with unsafe settings;
- HTTPS, security headers, a Docker image and Compose file;
- **email confirmation** at sign-up;
- a **durable email queue** for order and payment emails;
- debt reminders;
- **image uploads** to disk or S3, with GPS metadata removed;
- **coupon codes**;
- an **admin dashboard**;
- **backup and restore** scripts.

## Step 5: the frontend restructure (teammate)

The React app was reorganised and given an API layer, but it was written
against its own idea of the backend.

## Step 6: connecting the two halves

The frontend and backend did not fit: different field names, missing
endpoints and different assumptions. The fix:

- The frontend got one **API layer** (`src/api/`) with **mappers** that
  translate the backend's field names into the names the components already
  used. The existing screens did not need to be rewritten.
- The backend gained what the admin panel needed: categories as records,
  store settings, and admin user management (migration 000006).

## Step 7: many stores (multi-tenancy), in four parts

The goal: one codebase and one deployment serving many completely independent
stores. A plan was written and approved before any code changed.

1. **Stores and isolation** (migration 000007): a `stores` table; `store_id`
   on every table; the store found from the host name; isolation enforced by
   the database itself (row-level security).
2. **Platform super admin** (migration 000008): a separate kind of admin who
   creates stores, with an invite email for the store's admin; emails carry
   each store's own name and links.
3. **Store customization**: banners, feature tiles, footer and images stored
   per store, with the template's texts as defaults.
4. **Custom domains** (migration 000009): DNS verification, HTTPS
   certificates on demand, and one upload folder per store.

See [05-multi-tenancy.md](05-multi-tenancy.md), [14-admin-features.md](14-admin-features.md)
and [15-platform-admin.md](15-platform-admin.md).

## Step 8: Docker test and a real backup run

- The Docker images were built and the whole stack was tested on a machine
  where Docker runs: HTTPS, the React app on every route, the platform admin,
  creating a store, isolation between stores, uploads kept across restarts.
- Two fixes came out of it:
  - Caddy needs `ACME_EMAIL`, so it is now required.
  - The backup script now writes the uploads archive to a temporary file
    first, and turns off Git Bash's path rewriting on Windows.
- The backup scripts were run for real: a backup, a restore test into a
  separate database, and a live restore.

## Step 9: the missing frontend screens

The backend already had coupons, offers, bundles, pay later and the waiting
list, but the React app showed none of them. The screens were added:

- **Customers:**
  - a coupon box and bundle lines in the cart, with totals from the backend;
  - "Pay Later" at checkout;
  - "Notify me" on sold-out products;
  - Offers and Bundles pages;
  - My Balance and Waiting List tabs in the profile.
- **Store admins:** Coupons, Offers, Bundles, Customer Debts and Waiting List
  pages.

A new test (`TestSalesScreens`) sends exactly what these screens send. It
caught one real bug before release: offers require start and end dates, but
the first version of the form sent none.

## Step 10: the security and optimization checkup

A full review. The findings and fixes are in
[23-security-and-quality-review.md](23-security-and-quality-review.md). In
short:

- security headers for the React pages;
- broken email links fixed;
- product pages now show offer prices;
- a flaky test fixed;
- the admin panel split out of the shoppers' download (824 KB → 582 KB);
- the images compressed (9.0 MB → 0.55 MB; the whole build 10.7 MB → 1.7 MB);
- every test passes, also under Go's race detector.

## Step 11: shipping fees, the Emails page and CI

- **Shipping fees** (migration 000010): each store sets a standard fee,
  per-city fees and a "free from" amount. The backend charges the fee at
  checkout, the cart shows the same fee, and orders keep a copy. See
  [10-cart-checkout-orders.md](10-cart-checkout-orders.md) §10.3.
- **An Emails page for store admins:** it shows whether emails go out, and
  failed emails can be retried. See [13-emails.md](13-emails.md) §13.5.
- **GitHub CI:** every push runs the tests, lint, build, a vulnerability
  scan and the Docker builds. See [22-testing.md](22-testing.md) §22.4.

## Step 12: quick wins and small features

- **Cleanup:**
  - `npm audit fix` (0 vulnerabilities);
  - 14 unused images deleted;
  - `staticcheck` added to CI;
  - a missing "Pay Later" label fixed in customers' order details.
- **Private order notes for the store's team** (migration 000011).
- **A printable invoice / packing slip.**
- **CSV exports of orders and customers**, built by the backend for Excel
  (Arabic, store time, formula-safe).
- **Privacy Policy and Terms pages**, linked from the footer. They describe
  what the platform really does with data; they should be reviewed legally.

