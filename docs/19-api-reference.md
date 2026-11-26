# 19. API reference

Taken from `backend/internal/routes/routes.go`, which is the authority.
Reading that file shows the same list with the handler for each line.

**Access levels:**

- **P** public;
- **S** public with the strict account rate limit;
- **U** logged in;
- **UL** logged in + strict limit;
- **V** logged in with a confirmed email;
- **A** store admin;
- **PA** platform admin (only on the platform admin host).

**Conventions:**

- JSON bodies, snake_case fields.
- Money: JSON numbers with up to 2 decimals.
- Dates: RFC 3339 with time zone, e.g. `2026-12-01T00:00:00+03:00`.
- Lists: `?limit=` (1–100, default 20) and `?offset=` give
  `{items, total, limit, offset}`.

**Status codes:**

| Code | Meaning |
|---|---|
| 400 | bad input (see `fields`) |
| 401 | not logged in, expired, revoked or deactivated |
| 403 | not allowed, email not confirmed for checkout, or store suspended |
| 404 | not found, or not yours |
| 409 | not possible in the current state |
| 429 | too many requests (`Retry-After`) |
| 500 | unexpected (details only in the log) |

## Health and files

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/`, `/api/health` | P | health also pings the database |
| GET | `/uploads/{store}/{name}`, `/uploads/{name}` | P | uploaded images (local storage; the second form is for files from before per-store folders) |
| GET | `/internal/tls-check?domain=` | P (blocked from outside by Caddy) | Caddy asks before getting a certificate |

## Account

| Method | Path | Access | Body / notes |
|---|---|---|---|
| POST | `/api/register` | S | `first_name, last_name, email, password (8–72), phone` → 201 user; sends confirmation |
| POST | `/api/login` | S | `email, password` → `{user, token, expires_at}` |
| POST | `/api/email/verify` | S | `token` |
| POST | `/api/email/verify/resend` | U | new link (3 per hour) |
| POST | `/api/password/forgot` | S | `email` → always the same message |
| POST | `/api/password/reset` | S | `token, new_password` |
| POST | `/api/logout-all` | U | revoke every token |
| GET / PUT | `/api/profile` | U | PUT: `first_name, last_name, phone, profile_image_url?` |
| PUT | `/api/profile/password` | U | `current_password, new_password` → `{message, token, expires_at}` |
| POST / DELETE | `/api/profile/image` | U | multipart field `image` / remove |
| DELETE | `/api/profile` | U | deactivate the account |

## Catalog

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/products` | P | `?category= &search= &sort=newest\|price_asc\|price_desc\|name &in_stock= &on_offer=` |
| GET | `/api/products/{id}` and `/images`, `/colors`, `/reviews` | P | reviews: `?sort=newest\|helpful\|rating_high\|rating_low` |
| GET | `/api/categories` | P | active categories |
| GET | `/api/settings` | P | this store's name, logo, colors, content |
| GET | `/api/admin/products`, `/api/admin/products/{id}` | A | includes inactive |
| POST / PUT / DELETE | `/api/products`, `/api/products/{id}` | A | body may include `images: [url...]` and `colors: [...]`; PUT with stock > 0 emails the waiting list; DELETE deactivates |
| POST | `/api/products/{id}/images` | A | by URL |
| POST | `/api/admin/products/{id}/images/upload` | A | multipart `image`, `is_primary?`, `display_order?` |
| PUT | `/api/product-images/{id}/primary` | A | |
| DELETE | `/api/product-images/{id}` | A | also deletes an uploaded file |
| POST | `/api/products/{id}/colors` | A | `name, name_ar, hex_code` |
| DELETE | `/api/product-colors/{id}` | A | |
| GET / POST | `/api/admin/categories` | A | GET includes switched-off ones; POST `{name, name_ar, icon, is_active?}` |
| PUT / DELETE | `/api/admin/categories/{id}` | A | rename copies to products; DELETE hides its products |

## Cart, favorites, addresses

| Method | Path | Access | Notes |
|---|---|---|---|
| GET / DELETE | `/api/cart` | U | `{items, bundles, subtotal, coupon, discount, total, item_count, ...}`; DELETE empties it |
| POST | `/api/cart/items` | U | `product_id, quantity?, selected_color?` |
| PUT / DELETE | `/api/cart/items/{id}` | U | PUT `{quantity}` |
| POST | `/api/cart/bundles` | U | `bundle_id, quantity?` → the cart |
| PUT / DELETE | `/api/cart/bundles/{id}` | U | PUT `{quantity}` |
| POST | `/api/cart/coupon` | UL | `code` → the cart; 400 with the reason on `fields.code` |
| DELETE | `/api/cart/coupon` | U | |
| GET / POST | `/api/favorites` | U | POST `product_id` |
| DELETE | `/api/favorites/{product_id}` | U | |
| GET / POST | `/api/addresses` | U | |
| GET / PUT / DELETE | `/api/addresses/{id}` | U | |
| PUT | `/api/addresses/{id}/default` | U | |

## Orders

| Method | Path | Access | Notes |
|---|---|---|---|
| POST | `/api/orders` | V | `address_id?, payment_method? (cash_on_delivery \| pay_later)` |
| GET | `/api/orders`, `/api/orders/{id}` | U | your own only |
| POST | `/api/orders/{id}/cancel` | U | while pending |
| GET | `/api/admin/orders`, `/api/admin/orders/{id}` | A | `?status= &payment_method=` |
| PUT | `/api/admin/orders/{id}/status` | A | `status, payment_collected?` |

## Reviews and waiting list

| Method | Path | Access | Notes |
|---|---|---|---|
| POST | `/api/products/{id}/reviews` | U | `rating, comment?, comment_ar?` |
| PUT / DELETE | `/api/reviews/{id}` | U (author) | |
| POST / DELETE | `/api/reviews/{id}/helpful` | U | |
| POST | `/api/waiting-list` | U | `product_id, quantity?, selected_color?`; only out-of-stock products |
| GET | `/api/waiting-list` | U | your requests (a plain array) |
| DELETE | `/api/waiting-list/{id}` | U | cancel |
| GET | `/api/admin/waiting-list` | A | `?status=waiting\|notified\|fulfilled\|cancelled`; items include `customer` |
| PUT | `/api/admin/waiting-list/{id}/status` | A | `{status}`; `notified` sends the email |

## Offers, bundles, coupons

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/offers` | P | live offers |
| GET | `/api/offers/{id}` | P | `{offer, products: page}` |
| GET / POST | `/api/admin/offers` | A | `?status=`; POST `{name, name_ar, description, description_ar, discount_type, discount_value, scope, product_ids, category, starts_at, ends_at, is_active}` (both dates required) |
| GET / PUT / DELETE | `/api/admin/offers/{id}` | A | |
| GET | `/api/bundles`, `/api/bundles/{id}` | P | live bundles with computed fields |
| GET / POST | `/api/admin/bundles` | A | POST `{name, name_ar, description, description_ar, image_url, price, starts_at, ends_at, is_active, items: [{product_id, quantity, selected_color}]}` |
| GET / PUT / DELETE | `/api/admin/bundles/{id}` | A | DELETE deactivates |
| GET / POST | `/api/admin/coupons` | A | `?status=live\|scheduled\|expired\|disabled\|used_up &search=`; POST `{code, description, description_ar, discount_type, discount_value, scope, product_ids, category, min_order_amount, max_discount, usage_limit, per_customer_limit, starts_at, ends_at, is_active}` |
| GET / PUT / DELETE | `/api/admin/coupons/{id}` | A | DELETE keeps order history |

## Debts

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/debt` | U | `{account, transactions: page}` |
| GET | `/api/admin/debts` | A | `?outstanding=true` |
| GET | `/api/admin/customers/{id}/debt` | A | statement |
| PUT | `/api/admin/customers/{id}/credit-limit` | A | `{credit_limit}` (0 = off) |
| POST | `/api/admin/customers/{id}/debt/payments` | A | `{amount, note?, order_id?}` |
| POST | `/api/admin/customers/{id}/debt/adjustments` | A | `{amount (signed), note, order_id?}` |

## Store administration

| Method | Path | Access | Notes |
|---|---|---|---|
| GET | `/api/admin/dashboard` | A | `?days=` (1–366, default 30) |
| GET / POST | `/api/admin/users` | A | GET `?search= &role=`; POST `{first_name, last_name, email, phone?, password, role?, is_active?}` |
| GET / PUT / DELETE | `/api/admin/users/{id}` | A | GET `{user, orders, favorites}`; empty password keeps it; safety rules in [14](14-admin-features.md) |
| PUT | `/api/admin/settings` | A | `{storeName, logo, colors: {light, dark}, content: {hero, productsBanner, features, footer}}` |
| POST | `/api/admin/settings/images` | A | multipart `image` → `{url}` |
| GET / POST | `/api/admin/domains` | A | POST `{domain}` → the domain with `dns: {txt_name, txt_value, cname_target}` |
| POST | `/api/admin/domains/{id}/verify`, `/primary` | A | |
| DELETE | `/api/admin/domains/{id}` | A | |
| GET | `/api/admin/emails` | A | `?status=pending\|failed\|sent &kind= &to=` |
| GET | `/api/admin/emails/summary` | A | |
| GET / DELETE | `/api/admin/emails/{id}` | A | |
| POST | `/api/admin/emails/{id}/retry`, `/api/admin/emails/retry-failed` | A | |

## Platform (only on the platform admin host)

| Method | Path | Access | Notes |
|---|---|---|---|
| POST | `/api/platform/login` | S | `{email, password}` → `{admin, token, expires_at}` |
| GET | `/api/platform/me` | PA | |
| GET / POST | `/api/platform/stores` | PA | GET `?search= &status=`; POST `{slug, name, currency?, timezone?, admin: {first_name, last_name, email}}` |
| GET / PUT | `/api/platform/stores/{id}` | PA | PUT `{name, currency, timezone}` |
| POST | `/api/platform/stores/{id}/suspend`, `/activate`, `/invite` | PA | |
