# 17. Database and migrations

## 17.1 How migrations work

Migrations are numbered SQL files in `backend/migrations/`, each with an
`up` and a `down` file.

```
go run ./cmd/migrate status   # read-only: applied and pending
go run ./cmd/migrate up       # apply what is pending
```

- Each migration runs in **one transaction**: if anything fails, nothing
  changes.
- Applied migrations are recorded in `schema_migrations`, so each runs once.
- A database whose baseline was created by hand is recognised; only newer
  migrations run.
- With Docker Compose, the `migrate` service runs `up` **before** the API
  starts, on every deploy.
- `down` files exist to undo a migration by hand in an emergency. Back up
  first.

**Why versioned SQL files?** The database structure changes together with
the code, in the same commit, and every environment (laptop, test database,
server) gets exactly the same structure in the same order.

## 17.2 The migrations

| # | Adds | Why |
|---|---|---|
| 000001 baseline | users, products, product_images, product_colors, carts, cart_items, favorites, addresses, orders, order_items, reviews, review_helpful_votes, waiting_list_items | the original shop |
| 000002 | offers, offer_products, bundles, bundle_items, cart_bundles, order_bundles, order_bundle_items, users.credit_limit, customer_debt_transactions, `pay_later`, `deferred` | promotions and pay later |
| 000003 | users.token_version, password_reset_tokens | revocable logins, password reset |
| 000004 | users.email_verified_at (existing users set to verified), users.debt_reminded_at, email_verification_tokens, email_outbox | trusted emails, reliable notifications |
| 000005 | coupons, coupon_products, coupon_redemptions, carts.coupon_id, orders.subtotal_amount / discount_amount / coupon_id / coupon_code (+ check total = subtotal − discount), dashboard indexes | coupon codes, a fast dashboard |
| 000006 | categories (seeded from the products' categories), store_settings | what the admin panel manages |
| 000007 | stores, store_domains; store_id on every table (default = current store); store-aware foreign keys; per-store unique rules (email, category name, coupon code); row-level security; role stockwise_app; existing data = store 1 `main` | many stores in one database |
| 000008 | platform_admins; email_outbox.from_name | the platform super admin, per-store sender names |
| 000009 | store_domains.verify_token; one main domain per store | custom domains |

## 17.3 How the code respects the schema

- **Lengths:** validators check `varchar` lengths first, so the user gets a
  clear 400 naming the field instead of a database error.
- **Money:** decimals match `numeric(10,2)`.
- **Uniqueness:** unique indexes enforce one default address, one primary
  image, one review per product and one vote per review, and they get
  friendly 409 messages.
- **Last safety net:** `CHECK` constraints repeat the important rules
  (discount ranges, scope ⇔ category, debt signs, order total = subtotal −
  discount), so bad data cannot be stored even by hand.
- **`ON DELETE RESTRICT` on orders** is why accounts and products are
  deactivated rather than deleted.
- **Unique per store, not globally:** since 000007 an email, a category
  name or a coupon code is unique **per store**, so two shops can both have
  a customer `sara@example.com` or a coupon `SAVE10`.

## 17.4 Roles

| Role | Used by | Sees |
|---|---|---|
| the owner (e.g. `shop_api`) | migrations, backups | everything (owners skip RLS) |
| `stockwise_app` | the running API (it switches to it after connecting) | only the current store's rows, or all rows in platform mode |

**Why two roles?** If the API ran as the owner, PostgreSQL would skip the
isolation policies and they would protect nothing.
