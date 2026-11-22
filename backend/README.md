# E-commerce Backend (Go + PostgreSQL)

REST API for the StockWise shop: accounts, a bilingual (English/Arabic) product
catalog, cart, favorites, addresses, cash-on-delivery checkout, order tracking,
reviews with helpful votes, a waiting list for out-of-stock products, timed
**offers**, product **bundles**, and **pay-later customer debts**.

- How the code is organised and why: [docs/architecture.md](docs/architecture.md)
- Offers, bundles and debts in detail: [docs/offers-bundles-debts.md](docs/offers-bundles-debts.md)
- Rate limiting, login revocation and password reset: [docs/auth-security.md](docs/auth-security.md)

## Requirements

- Go 1.27+
- PostgreSQL 18 (the schema was dumped from 18.6)

## Setup

1. Create an empty PostgreSQL database (e.g. `ecommerce`).
2. Copy `.env.example` to `.env` and fill it in. `JWT_SECRET` must be at least
   32 characters; the server refuses to start otherwise.
3. Apply the schema with the built-in migration tool (no `psql` needed):

   ```powershell
   go run ./cmd/migrate status   # read-only: what is applied / pending
   go run ./cmd/migrate up       # apply pending migrations
   ```

   It records applied migrations in `schema_migrations`. A database whose
   baseline was applied by hand earlier is recognised, so only newer
   migrations run. **Run `up` after every pull that adds a migration.**
4. Run the API:

   ```powershell
   go run ./cmd/api
   ```

5. Check `GET http://localhost:8080/api/health` → `{"database":"ok","status":"ok"}`.

### Creating the first admin

Registration always creates customers. Promote an account once with SQL:

```sql
UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
```

## Tests

```powershell
go test ./...                      # unit tests, no database needed

# end-to-end tests: creates a throwaway database, runs the real API, drops it
$env:INTEGRATION_DATABASE_URL = "postgres://USER:PASS@localhost:5432/postgres?sslmode=disable"
go test ./tests -v
```

## Conventions for the frontend

- **Auth:** `POST /api/login` returns `{user, token, expires_at}`. Send
  `Authorization: Bearer <token>` on protected routes. A 401 saying the login
  was revoked means the token was invalidated (password change/reset or "log
  out everywhere"): delete it and show the login page.
- **Errors** are always JSON: `{"error": "message"}`. Validation errors also
  carry per-field messages: `{"error": "validation failed", "fields": {"email": "is required"}}`.

  | Status | Meaning |
  |---|---|
  | 400 | Bad input (see `fields`) |
  | 401 | Not logged in, bad/expired token, or deactivated account |
  | 403 | Logged in but not an admin |
  | 404 | Not found (or not yours) |
  | 409 | Valid request but the current state forbids it (out of stock, already reviewed, wrong order status, ...) |
  | 429 | Rate limit reached - wait `Retry-After` seconds |
  | 500 | Unexpected server error (details are only in the server log) |

- **Money** is sent as JSON numbers with at most 2 decimals (e.g. `19.99`).
  `discount_value` is a percentage. `final_price` / `unit_price` already include the discount.
- **Lists** that can grow are paginated: `?limit=` (1-100, default 20) and
  `?offset=`; the response is `{items, total, limit, offset}`.
- **Cart** endpoints always return the full updated cart.
- **CORS:** allowed origins come from `CORS_ALLOWED_ORIGINS` (default `http://localhost:5173`).

## API reference

Access: **P** public · **U** logged-in user · **A** admin.

### Health & account

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/` | P | Process is up |
| GET | `/api/health` | P | Also pings the database |
| POST | `/api/register` | P | `first_name, last_name, email, password (8-72), phone, profile_image_url?` (rate-limited) |
| POST | `/api/login` | P | `email, password` (email is case-insensitive; rate-limited per IP and per email) |
| POST | `/api/password/forgot` | P | `email` - always the same reply; emails a reset link if the account exists |
| POST | `/api/password/reset` | P | `token, new_password` - token from the emailed link; logs out every device |
| POST | `/api/logout-all` | U | Revokes every token of the account, including this one |
| GET | `/api/profile` | U | |
| PUT | `/api/profile` | U | `first_name, last_name, phone, profile_image_url?` |
| PUT | `/api/profile/password` | U | `current_password, new_password` → `{message, token, expires_at}`: other devices are logged out, use the new token |
| DELETE | `/api/profile` | U | Deactivates the account (soft delete) |

### Catalog

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/products` | P | `?category= &search= &sort=newest\|price_asc\|price_desc\|name &in_stock=true &on_offer=true &limit= &offset=`; `final_price` and price sorting include live offers |
| GET | `/api/products/{id}` | P | Includes `images` and `colors` |
| GET | `/api/products/{id}/images` | P | |
| GET | `/api/products/{id}/colors` | P | |
| GET | `/api/categories` | P | Distinct categories with product counts |
| GET | `/api/admin/products` | A | Same filters, includes inactive products |
| GET | `/api/admin/products/{id}` | A | Also inactive products |
| POST | `/api/products` | A | `name, name_ar, description, description_ar, category, category_ar, price, discount_value, stock, is_active?` |
| PUT | `/api/products/{id}` | A | Same body; omitted `is_active` keeps the current value |
| DELETE | `/api/products/{id}` | A | Soft delete (`is_active=false`) |
| POST | `/api/products/{id}/images` | A | `image_url, is_primary, display_order`; first image becomes primary |
| PUT | `/api/product-images/{id}/primary` | A | Make this the primary image |
| DELETE | `/api/product-images/{id}` | A | Next image is promoted if the primary is deleted |
| POST | `/api/products/{id}/colors` | A | `name, name_ar, hex_code (#RRGGBB)` |
| DELETE | `/api/product-colors/{id}` | A | |

### Cart, favorites, addresses

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/cart` | U | `items` (single products, with `offer`) and `bundles`; live prices, `line_total`, `is_available`, `unavailable_reason`, `subtotal` |
| POST | `/api/cart/items` | U | `product_id, quantity (default 1), selected_color` (required if the product has colors) |
| PUT | `/api/cart/items/{id}` | U | `quantity, selected_color?` |
| DELETE | `/api/cart/items/{id}` | U | |
| DELETE | `/api/cart` | U | Empty the cart (products and bundles) |
| POST | `/api/cart/bundles` | U | `bundle_id, quantity (default 1)` |
| PUT | `/api/cart/bundles/{id}` | U | `quantity` |
| DELETE | `/api/cart/bundles/{id}` | U | |
| GET | `/api/favorites` | U | Each favorite embeds the product |
| POST | `/api/favorites` | U | `product_id` → 201 new / 200 already saved |
| DELETE | `/api/favorites/{product_id}` | U | |
| GET | `/api/addresses` | U | Default first |
| POST | `/api/addresses` | U | `recipient_name, phone, city, area, street, address_label?, building_number?, apartment_number?, additional_instructions?, is_default?`; first address becomes default |
| GET / PUT / DELETE | `/api/addresses/{id}` | U | Deleting the default promotes the newest remaining address |
| PUT | `/api/addresses/{id}/default` | U | |

### Orders

| Method | Path | Access | Notes |
|---|---|---|---|
| POST | `/api/orders` | U | `{address_id?, payment_method?}` - default address if omitted; `payment_method` = `cash_on_delivery` (default) or `pay_later`. Checks out the whole cart, bundles included |
| GET | `/api/orders` | U | Own orders with items |
| GET | `/api/orders/{id}` | U | |
| POST | `/api/orders/{id}/cancel` | U | Only while `pending`; stock is restored |
| GET | `/api/admin/orders` | A | `?status= &payment_method= &limit= &offset=`, includes customer |
| GET | `/api/admin/orders/{id}` | A | |
| PUT | `/api/admin/orders/{id}/status` | A | `{status, payment_collected?}` - see the life cycle below |

Order life cycle (anything else is refused with 409):

```
pending → confirmed → processing → shipped → delivered
   └──────────┴────────────┴──→ cancelled   (stock is put back, debt reversed)
```

Delivering an order marks the customer's reviews of its products (including
products inside bundles) as verified purchases, and settles the payment:
cash orders become `paid`, or `deferred` (moved to the customer's debt) when
sent with `payment_collected: false`; pay-later orders stay `deferred`.

### Reviews

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/products/{id}/reviews` | P | `?sort=newest\|helpful\|rating_high\|rating_low` |
| POST | `/api/products/{id}/reviews` | U | `comment, comment_ar?, rating (1-5)`; one per product |
| PUT | `/api/reviews/{id}` | Author | |
| DELETE | `/api/reviews/{id}` | Author | |
| POST | `/api/reviews/{id}/helpful` | U | Once per review, not on your own |
| DELETE | `/api/reviews/{id}/helpful` | U | |

### Waiting list

| Method | Path | Access | Notes |
|---|---|---|---|
| POST | `/api/waiting-list` | U | `product_id, quantity?, selected_color?` - only for out-of-stock products |
| GET | `/api/waiting-list` | U | |
| DELETE | `/api/waiting-list/{id}` | U | Cancels the request |
| GET | `/api/admin/waiting-list` | A | `?status= &product_id= &limit= &offset=`, oldest first |
| PUT | `/api/admin/waiting-list/{id}/status` | A | `waiting ⇄ notified → fulfilled / cancelled` |

Marking a request `notified` records `notified_at`; it does not send an email
or SMS - that needs a notification provider.

### Offers

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/offers` | P | Live offers |
| GET | `/api/offers/{id}` | P | `{offer, products}` - a live offer and the products it covers |
| GET | `/api/admin/offers` | A | `?status=live\|scheduled\|expired\|disabled` |
| GET | `/api/admin/offers/{id}` | A | |
| POST | `/api/admin/offers` | A | `name, name_ar, description?, description_ar?, discount_type (percentage\|fixed), discount_value, scope (products\|category), product_ids \| category, starts_at, ends_at, is_active?` |
| PUT | `/api/admin/offers/{id}` | A | Same body |
| DELETE | `/api/admin/offers/{id}` | A | Orders keep the offer name |

The customer always gets the single best price (product discount or one
offer); discounts never stack.

### Bundles

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/bundles` | P | Live bundles with `items`, `original_total`, `savings`, `available_stock` |
| GET | `/api/bundles/{id}` | P | |
| GET | `/api/admin/bundles` | A | Including inactive / outside their window |
| GET | `/api/admin/bundles/{id}` | A | |
| POST | `/api/admin/bundles` | A | `name, name_ar, description?, description_ar?, image_url?, price, starts_at?, ends_at?, is_active?, items: [{product_id, quantity, selected_color?}]` |
| PUT | `/api/admin/bundles/{id}` | A | Same body (replaces the items) |
| DELETE | `/api/admin/bundles/{id}` | A | Soft delete |

### Customer debts (pay later)

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/debt` | U | `{account, transactions}` - my balance, credit and history |
| GET | `/api/admin/debts` | A | `?outstanding=true`; biggest balance first |
| GET | `/api/admin/customers/{id}/debt` | A | One customer's statement |
| PUT | `/api/admin/customers/{id}/credit-limit` | A | `{credit_limit}` (0 = pay later off) |
| POST | `/api/admin/customers/{id}/debt/payments` | A | `{amount, note?, order_id?}` |
| POST | `/api/admin/customers/{id}/debt/adjustments` | A | `{amount (+/-), note, order_id?}` |

## Not included yet

- Coupon codes typed at checkout (offers are automatic only)
- Online payments (cash on delivery and pay later only)
- Image upload (the API stores image URLs)
- Email/SMS notifications other than password reset (order updates, waiting
  list, debt reminders) - the `mail` package is ready to be reused
- Email verification at sign-up, CAPTCHA
- Shared rate-limit storage (Redis) for running several API instances
