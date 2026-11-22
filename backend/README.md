# E-commerce Backend (Go + PostgreSQL)

REST API for the StockWise shop: accounts, a bilingual (English/Arabic) product
catalog, cart, favorites, addresses, cash-on-delivery checkout, order tracking,
reviews with helpful votes, and a waiting list for out-of-stock products.

For how the code is organised and why, read [docs/architecture.md](docs/architecture.md).

## Requirements

- Go 1.27+
- PostgreSQL 18 (the schema was dumped from 18.6)

## Setup

1. Create an empty database and apply the schema:

   ```powershell
   psql -U postgres -c "CREATE DATABASE ecommerce"
   psql -U postgres -d ecommerce -f migrations/000001_baseline.up.sql
   ```

   Only run the migration on a **new, empty** database.
2. Copy `.env.example` to `.env` and fill it in. `JWT_SECRET` must be at least
   32 characters; the server refuses to start otherwise.
3. Run the API:

   ```powershell
   go run ./cmd/api
   ```

4. Check `GET http://localhost:8080/api/health` → `{"database":"ok","status":"ok"}`.

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
  `Authorization: Bearer <token>` on protected routes.
- **Errors** are always JSON: `{"error": "message"}`. Validation errors also
  carry per-field messages: `{"error": "validation failed", "fields": {"email": "is required"}}`.

  | Status | Meaning |
  |---|---|
  | 400 | Bad input (see `fields`) |
  | 401 | Not logged in, bad/expired token, or deactivated account |
  | 403 | Logged in but not an admin |
  | 404 | Not found (or not yours) |
  | 409 | Valid request but the current state forbids it (out of stock, already reviewed, wrong order status, ...) |
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
| POST | `/api/register` | P | `first_name, last_name, email, password (8-72), phone, profile_image_url?` |
| POST | `/api/login` | P | `email, password` (email is case-insensitive) |
| GET | `/api/profile` | U | |
| PUT | `/api/profile` | U | `first_name, last_name, phone, profile_image_url?` |
| PUT | `/api/profile/password` | U | `current_password, new_password` |
| DELETE | `/api/profile` | U | Deactivates the account (soft delete) |

### Catalog

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/products` | P | `?category= &search= &sort=newest\|price_asc\|price_desc\|name &in_stock=true &limit= &offset=` |
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
| GET | `/api/cart` | U | Live prices, `line_total`, `is_available`, `unavailable_reason`, `subtotal` |
| POST | `/api/cart/items` | U | `product_id, quantity (default 1), selected_color` (required if the product has colors) |
| PUT | `/api/cart/items/{id}` | U | `quantity, selected_color?` |
| DELETE | `/api/cart/items/{id}` | U | |
| DELETE | `/api/cart` | U | Empty the cart |
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
| POST | `/api/orders` | U | `{address_id?}` - uses the default address if omitted. Checks out the whole cart |
| GET | `/api/orders` | U | Own orders with items |
| GET | `/api/orders/{id}` | U | |
| POST | `/api/orders/{id}/cancel` | U | Only while `pending`; stock is restored |
| GET | `/api/admin/orders` | A | `?status= &limit= &offset=`, includes customer |
| GET | `/api/admin/orders/{id}` | A | |
| PUT | `/api/admin/orders/{id}/status` | A | `{status}` - see the life cycle below |

Order life cycle (anything else is refused with 409):

```
pending → confirmed → processing → shipped → delivered
   └──────────┴────────────┴──→ cancelled   (stock is put back)
```

Delivering an order sets `payment_status` to `paid` (cash collected) and marks
the customer's reviews of those products as verified purchases.

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

## Not included yet

- Online payments (the schema only allows cash on delivery)
- Image upload (the API stores image URLs)
- Email/SMS sending (password reset, waiting-list notifications)
- Logging out other devices after a password change: tokens stay valid until
  they expire (`JWT_TTL`). Fixing this needs a new column such as
  `users.password_changed_at`.
- Rate limiting on login/register (do this at the reverse proxy)
