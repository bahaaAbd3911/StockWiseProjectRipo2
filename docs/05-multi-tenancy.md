# 5. Many stores in one system (multi-tenancy)

"Multi-tenant" means one installation serves many customers of the platform
(here: stores), each seeing only its own data. This is the most important
design in the project, because a mistake here would show one shop's
customers and orders to another shop.

## 5.1 The choice: one database, shared tables, `store_id` on every row

| Option | Isolation | Cost to run | Adding a store | Chosen |
|---|---|---|---|---|
| a database per store | strongest | high: N databases to migrate, back up and connect to | create a database | no |
| a schema per store | strong | medium: N schemas to migrate | create a schema | no |
| **shared tables + `store_id` + row-level security** | strong, **enforced by PostgreSQL** | **low: one database, one migration, one backup** | **insert one row** | **yes** |

**Why:** the requirement was "one codebase, one deployment, a store ready
without a new database". Shared tables meet that directly. Their usual
weakness, a forgotten `WHERE store_id = ...` leaking data, is removed by
making **the database** enforce the filter (5.3), so a bug in the Go code
cannot leak data.

## 5.2 Finding the store from the host name

`internal/tenant` (the resolver) and `middleware.Store`:

| Host | Store |
|---|---|
| `<slug>.PLATFORM_DOMAIN`, e.g. `ahmad.myplatform.com` | the store with slug `ahmad` |
| a **verified** custom domain, e.g. `ahmadstore.com` | the store that owns it |
| `PLATFORM_ADMIN_HOST` (default `admin.<PLATFORM_DOMAIN>`) | no store: only the platform API |
| anything else (localhost, an IP, the bare domain) | `DEFAULT_STORE` (`main`), or 404 if that setting is empty |

- Unknown store → **404**; suspended store → **403** "this store is
  currently unavailable".
- `/api/health`, `/uploads/...` and `/internal/tls-check` work on any host.
- Answers are cached for 30 seconds. Suspending a store clears its cache
  entry at once.

**Why the host name, and not a header, a path like `/store/ahmad`, or a
`store_id` in the body?**

- The host is what the visitor typed. It cannot be changed by editing a
  request without also going to a different address.
- Browsers keep logins (localStorage) **per host**, so a login in one store
  is not even visible to another store's pages.
- It works the same way for subdomains and custom domains.
- **The store is never taken from anything the client sends.** There is no
  `store_id` in any body, header or query string, so it cannot be forged.

## 5.3 Three layers of isolation

### Layer 1: row-level security (RLS)

- The API connects to PostgreSQL and then switches to the role
  `stockwise_app`. That role does **not own** the tables, so PostgreSQL
  applies its policies to it. Owners would skip them.
- Every store table has a policy named `store_isolation`:

  ```sql
  USING (app_is_platform() OR store_id = app_current_store())
  ```

- `app_current_store()` reads the connection setting `app.store_id`. A
  connection with **no store set sees no rows at all**, so a bug fails
  closed and never leaks.

### Layer 2: `store_id` fills itself in

The column's default is `app_current_store()`. All existing `INSERT`
statements kept working without passing a store, and inserting a row with
another store's id is refused by the policy.

### Layer 3: store-aware foreign keys

Every link includes the store:

```sql
FOREIGN KEY (store_id, product_id) REFERENCES products (store_id, id)
```

Foreign-key checks ignore RLS policies, so without this a cart line could
in theory point at another store's product. With it, the database refuses.

### How the connection knows the store

`database.ConnectApp` wraps every connection:

- Before each statement it compares the request's store (from the Go
  `context`) with the connection's current `app.store_id`, and sends
  `set_config(...)` only when they differ. Consecutive queries for the same
  store therefore cost nothing extra.
- A **transaction belongs to one store**: trying to switch store inside one
  is an error. A transaction mixing two stores would be exactly the bug this
  design prevents.

### Platform mode

Some trusted server code must see every store:

- the email worker;
- debt reminders and clean-up jobs;
- the platform admin's API.

`tenant.AsPlatform(ctx)` sets `app.platform = on` for those. **It can only
be switched on by Go code, never by anything in a request.**

## 5.4 Logins per store

- A login token records its store (the `sid` claim). On any other store's
  host it is refused: "this login belongs to another store".
- Customer accounts are per store. The same email can have an account in
  two stores, with different passwords. To a customer the stores are
  unrelated shops.
- Platform admins are in a separate table with their own token issuer. A
  store admin's token is useless on the platform API, and the other way
  round.
- Rate-limit counters (failed logins, reset emails) are per store too, so
  failures in one store never lock an account in another.

## 5.5 What is per store

Products, categories, offers, bundles, coupons, customers, carts,
favorites, addresses, orders, reviews, waiting lists, debts, emails in the
queue, dashboard numbers, store settings and page content, uploaded files
(`uploads/<store id>/`), custom domains, the email sender name, email
links and the currency.

## 5.6 Custom domains and HTTPS

1. The store admin adds `shop.example.com`. The API creates a random
   verification token.
2. At their DNS provider they add two records:
   - `TXT _stockwise-verify.shop.example.com "stockwise-verify=<token>"`,
     which proves they own the domain;
   - `CNAME shop.example.com → ahmad.myplatform.com`, which sends visitors
     to the platform.
3. They press **Verify**. The API looks up the TXT record. Only a verified
   domain shows the store.
4. On the first visit, Caddy asks the API `GET /internal/tls-check?domain=...`
   before getting a certificate. Only real hosts pass:
   - store subdomains;
   - verified domains of active stores;
   - the platform admin host;
   - the bare platform domain.

   Nobody can make the server request certificates for random names (that
   would quickly hit Let's Encrypt's limits). `/internal/*` is blocked from
   outside by the Caddyfile.
5. One verified domain can be the store's **main address**. Emails then
   link to it.

**Why a TXT record?** Pointing a domain at the server proves nothing:
anyone can point any domain anywhere. Publishing a secret token in that
domain's DNS can only be done by its owner.

## 5.7 How it is proven

The end-to-end tests try to break isolation on purpose. One store attempts
to read, change, delete or link to another store's data:

- through the API;
- with edited ids;
- with another store's token;
- **with raw SQL on the API's own database connection**: no store set →
  no rows; cross-store reads, writes and links are refused; a store switch
  inside a transaction is refused.

See [22-testing.md](22-testing.md) (`TestStoresAreIsolated`,
`TestDatabaseEnforcesIsolation`).

## 5.8 How the old single shop became store 1

Migration 000007 created the `stores` table and made every existing row
belong to store 1, `main`. A shop that existed before kept all its data and
kept working at its address. The migration's `down` file can only undo it
while `main` is the only store.
