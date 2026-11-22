# Offers, Bundles and Customer Debts

This document describes the three features added in migration
`000002_offers_bundles_debts`: what each one does, the rules it follows, the
database tables behind it, every API endpoint with examples, and the files
that implement it.

- [1. Before you start: applying the migration](#1-before-you-start-applying-the-migration)
- [2. Offers (timed promotions)](#2-offers-timed-promotions)
- [3. Bundles](#3-bundles)
- [4. Customer debts (pay later)](#4-customer-debts-pay-later)
- [5. How the three features meet at checkout](#5-how-the-three-features-meet-at-checkout)
- [6. Changes to existing endpoints](#6-changes-to-existing-endpoints)
- [7. Files added and changed](#7-files-added-and-changed)
- [8. Tests](#8-tests)
- [9. Limits and possible next steps](#9-limits-and-possible-next-steps)

---

## 1. Before you start: applying the migration

The new code needs the new tables. Apply them with the built-in migration tool
(no `psql` needed):

```powershell
go run ./cmd/migrate status   # read-only: shows applied / pending
go run ./cmd/migrate up       # applies the pending migrations
```

- The tool reads `DATABASE_URL` from `.env`.
- It records applied migrations in a `schema_migrations` table.
- A database created earlier by running `000001_baseline.up.sql` by hand is
  recognised: the baseline is marked as applied and only `000002` runs.
- Each migration runs in one transaction. If anything fails, nothing changes.
- `000002` only **adds** tables/columns, plus it widens two `CHECK`
  constraints on `orders` (new payment method and payment status). Existing
  data is untouched.
- Rollback: `migrations/000002_offers_bundles_debts.down.sql` (deletes all
  offers, bundles and debt records - back up first).

**Start the API only after migrating**, otherwise product and cart queries fail
because the new tables do not exist yet.

---

## 2. Offers (timed promotions)

### What it does

An admin creates a promotion with a start and end time. While it is running it
lowers the price of the products it covers - automatically, with no code for
the customer to type. When it ends, prices go back by themselves.

An offer has:

| Field | Meaning |
|---|---|
| `discount_type` | `percentage` (e.g. 20 = 20% off) or `fixed` (e.g. 15 = 15.00 off each unit) |
| `discount_value` | the percent (0 < v ≤ 100) or the amount (> 0) |
| `scope` | `products` = the products listed in `product_ids`; `category` = every product whose `category` matches (case-insensitive) |
| `starts_at`, `ends_at` | the window, RFC 3339 with time zone, e.g. `2026-12-01T00:00:00+03:00`; live while `starts_at ≤ now < ends_at` |
| `is_active` | an on/off switch to pause an offer without deleting it |
| `status` | computed: `live`, `scheduled` (not started), `expired` (ended), `disabled` (switched off) |

### The pricing rule: one best price, no stacking

A product can already have its own discount (`products.discount_value`), and
several offers can cover it at the same time. The customer always gets the
**single lowest price**; discounts are never added together.

```
candidates = [ price after the product's own discount,
               price after offer 1, price after offer 2, ... ]
final_price = the lowest candidate
offer       = the offer that produced it (null if the product discount won or tied)
```

Each candidate is calculated like this:

- percentage: `price × (100 − value) / 100`, rounded to cents
- fixed: `price − value`, never below 0

Example (price 100, own discount 10%):

| Live offers | final_price | offer |
|---|---|---|
| none | 90.00 | null |
| 25% sale | 75.00 | the 25% sale |
| 25% sale + 10.00 coupon (= 90) | 75.00 | the 25% sale (not 65 - no stacking) |
| only a 10.00 coupon (= 90, tie) | 90.00 | null - the product's own discount wins a tie |

This rule lives in exactly one function, `pricing.BestPrice`, and is used by
the product list, product detail, favorites, cart and checkout, so the price
a customer sees is the price they pay.

**Offers do not apply to bundles**: a bundle's price is already the deal.

### Where customers see it

- `final_price` and `offer {id, name, name_ar, discount_type, discount_value, ends_at}`
  on every product (lists, details, favorites, bundle contents).
- `unit_price` and `offer` on every cart line.
- `offer_id` and `offer_name` on every order item (copied at checkout, so the
  order keeps the name even if the offer is later deleted).
- `GET /api/products?on_offer=true` lists only products with a live offer, and
  `sort=price_asc|price_desc` sorts by the final price including offers.

### Endpoints

| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/api/offers` | Public | Live offers (paginated) |
| GET | `/api/offers/{id}` | Public | A live offer + a page of the products it covers: `{offer, products}` |
| GET | `/api/admin/offers?status=` | Admin | All offers; `status` = `live`, `scheduled`, `expired` or `disabled` |
| GET | `/api/admin/offers/{id}` | Admin | One offer, any status |
| POST | `/api/admin/offers` | Admin | Create |
| PUT | `/api/admin/offers/{id}` | Admin | Replace (send the full body; omitted `is_active` keeps the current value) |
| DELETE | `/api/admin/offers/{id}` | Admin | Delete permanently (orders keep the name). To pause, set `is_active: false` instead |

Create a product offer:

```json
POST /api/admin/offers
{
  "name": "Spring sale", "name_ar": "تخفيضات الربيع",
  "description": "25% off selected phones", "description_ar": "",
  "discount_type": "percentage", "discount_value": 25,
  "scope": "products", "product_ids": [12, 15, 19],
  "starts_at": "2026-03-01T00:00:00+03:00",
  "ends_at":   "2026-03-15T00:00:00+03:00"
}
```

Create a category offer (fixed amount):

```json
{
  "name": "Cases week", "name_ar": "أسبوع الأغطية",
  "discount_type": "fixed", "discount_value": 5,
  "scope": "category", "category": "Cases",
  "starts_at": "2026-04-01T00:00:00+03:00", "ends_at": "2026-04-08T00:00:00+03:00"
}
```

Response (shortened):

```json
{ "id": 3, "name": "Spring sale", "discount_type": "percentage", "discount_value": 25,
  "scope": "products", "category": null, "product_ids": [12, 15, 19],
  "starts_at": "...", "ends_at": "...", "is_active": true, "status": "scheduled" }
```

### Validation and errors

| Situation | Response |
|---|---|
| missing names, percentage > 100, value ≤ 0, more than 2 decimals | 400 with `fields` |
| `scope: products` without `product_ids`, or with a `category` | 400 |
| `scope: category` without `category`, or with `product_ids` | 400 |
| `ends_at` not after `starts_at` | 400 |
| a product ID that does not exist | 404 `products not found: [..]` |
| `GET /api/offers/{id}` for an offer that is not live | 404 |

### Database

```
offers          id, name, name_ar, description, description_ar,
                discount_type, discount_value, scope, category,
                starts_at, ends_at, is_active, created_at, updated_at
offer_products  offer_id → offers (CASCADE), product_id → products (CASCADE)
order_items     + offer_id → offers (SET NULL), + offer_name (snapshot)
```

`CHECK` constraints repeat the main rules (type, value range, scope ⇔
category, window order), so bad data cannot be stored even by hand.

---

## 3. Bundles

### What it does

A bundle is several products sold together as **one cart item** at its own
price, for example "Desk set = Keyboard (Black) + 2 × Mouse for 75.00".
Checkout takes the stock from every product inside it.

| Field | Meaning |
|---|---|
| `price` | what the customer pays for one bundle |
| `items` | `[{product_id, quantity, selected_color}]`, at least 2 products, or 1 product with quantity ≥ 2 ("3 for 2"); at most 20 |
| `selected_color` | required when the product has colors. The bundle fixes the color; the customer does not choose it |
| `starts_at`, `ends_at` | optional sale window (null = no limit) |
| `is_active` | on/off; `DELETE` sets it to false (soft delete) |

Computed for display:

| Field | Meaning |
|---|---|
| `original_total` | what the items would cost bought separately **today** (with their own best prices) |
| `savings` | `original_total − price`, never below 0 |
| `available_stock` | how many complete bundles the current stock allows: `min(stock ÷ quantity)` over its products |
| `is_available`, `unavailable_reason` | whether one bundle can be bought now, and why not |

A bundle is **live** when `is_active` is true and now is inside its window.
Customers only see live bundles.

### Availability rule (shared by bundle pages, cart and checkout)

A bundle line cannot be bought when:
1. the bundle is not live → "this bundle is no longer available"
2. a product inside it was deactivated → "X is no longer available"
3. the bundle's color for a product was deleted → "the color of X in this bundle is no longer available"
4. a product inside it gained colors after the bundle was made → "X in this bundle has no color set; please contact the shop"
5. there is not enough stock for `bundle quantity × item quantity` → "not enough stock of X"

### Bundles in the cart

The cart now has two lists: `items` (single products) and `bundles`.
`subtotal` adds both (available lines only) and `item_count` counts both.

```json
"bundles": [{
  "id": 4, "bundle_id": 2, "name": "Desk set", "unit_price": 75, "quantity": 1, "line_total": 75,
  "items": [{"product_id": 7, "product_name": "Keyboard", "quantity": 1, "selected_color": "Black", ...},
            {"product_id": 8, "product_name": "Mouse",    "quantity": 2, ...}],
  "is_available": true
}]
```

**Combined stock:** a product can be in the cart on its own *and* inside a
bundle. Each line can look fine alone while together they need more than the
stock. Checkout adds everything up per product and refuses with, for example:
`Mouse: only 4 left in stock, but your cart needs 5`.

### Orders

An order now has `bundles` next to `items`. Each bundle is a snapshot: name,
price, quantity, and the products inside it with their colors at that moment.
Changing or deactivating the bundle later never changes past orders.

- **Cancel** gives back the stock of every product inside every bundle.
- **Delivery** marks the customer's reviews of products bought inside a bundle
  as verified purchases (and future reviews count bundle purchases too).

### Endpoints

| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/api/bundles` | Public | Live bundles with items and computed fields |
| GET | `/api/bundles/{id}` | Public | One live bundle |
| GET | `/api/admin/bundles` | Admin | All bundles, including inactive / outside their window |
| GET | `/api/admin/bundles/{id}` | Admin | One bundle, any state |
| POST | `/api/admin/bundles` | Admin | Create |
| PUT | `/api/admin/bundles/{id}` | Admin | Replace, including the items list |
| DELETE | `/api/admin/bundles/{id}` | Admin | Deactivate |
| POST | `/api/cart/bundles` | User | `{bundle_id, quantity}` (quantity defaults to 1; adds to an existing line) |
| PUT | `/api/cart/bundles/{id}` | User | `{quantity}` |
| DELETE | `/api/cart/bundles/{id}` | User | Remove the bundle line |

The cart endpoints return the full updated cart, like the other cart endpoints.

Create:

```json
POST /api/admin/bundles
{
  "name": "Desk set", "name_ar": "طقم المكتب",
  "description": "Everything for your desk", "image_url": "https://cdn.example.com/desk.webp",
  "price": 75,
  "starts_at": null, "ends_at": null,
  "items": [
    {"product_id": 7, "quantity": 1, "selected_color": "Black"},
    {"product_id": 8, "quantity": 2}
  ]
}
```

### Errors

| Situation | Response |
|---|---|
| fewer than 2 products (and not one product ×2+), the same product twice, quantity outside 1-100 | 400 `fields.items` |
| a product with colors but no `selected_color`, or a color it does not have | 400 `fields.items` naming the product |
| unknown product ID | 404 |
| adding to the cart more bundles than the stock allows | 409 "not enough stock of X" |
| bundle not live | 404 on public GET / adding to the cart |

### Database

```
bundles             id, name, name_ar, description, description_ar, image_url, price,
                    starts_at, ends_at, is_active, created_at, updated_at
bundle_items        bundle_id → bundles (CASCADE), product_id → products (CASCADE),
                    quantity, selected_color      PRIMARY KEY (bundle_id, product_id)
cart_bundles        id, cart_id → carts (CASCADE), bundle_id → bundles (CASCADE),
                    quantity, added_at            UNIQUE (cart_id, bundle_id)
order_bundles       id, order_id → orders (CASCADE), bundle_id → bundles (SET NULL),
                    bundle_name, bundle_name_ar, quantity, unit_price, subtotal
order_bundle_items  id, order_bundle_id → order_bundles (CASCADE), product_id → products (SET NULL),
                    product_name, product_name_ar, quantity, selected_color, selected_color_ar
```

Bundles get their own cart table because `cart_items` requires a
`product_id` and allows one row per product.

---

## 4. Customer debts (pay later)

### What it does

Trusted customers can buy now and pay later. The shop keeps a ledger per
customer, like a debt notebook:

1. An admin gives a customer a **credit limit** (e.g. 500.00). The default is
   0, which means "pay later" is not available.
2. At checkout the customer chooses `payment_method: "pay_later"`. If
   `current balance + order total ≤ credit limit`, the order is placed, its
   `payment_status` is `deferred`, and the total is added to their debt.
3. The admin records money as the customer pays it back.
4. A cash order can also turn into debt: when the admin marks it delivered
   with `payment_collected: false` (the customer did not pay the driver).

### The ledger

Every change is one row in `customer_debt_transactions`. Rows are never
edited or deleted; mistakes are corrected with an adjustment row, so the
history is complete. **Balance = sum of all amounts.**

| type | sign | created when |
|---|---|---|
| `charge` | + | pay-later checkout (automatic); cash order delivered without payment (by the admin) |
| `payment` | − | admin records money received |
| `reversal` | − | a charged order is cancelled (automatic: removes what is still charged for that order) |
| `adjustment` | + or − | admin correction; a note is required |

Each transaction in the API also shows `balance_after`, the running balance
right after it, so the list reads like a bank statement.

Rules:
- A payment cannot be larger than the balance (the shop does not keep credit
  for customers).
- An adjustment cannot make the balance negative.
- A payment or adjustment can be linked to one of the customer's orders
  (`order_id`) for reference.
- Lowering a credit limit below the current balance is allowed; it only blocks
  new pay-later orders.
- Cancelling a pay-later order removes its remaining charge. Payments already
  made for it are not removed (that money was really received).
- Everything that changes a balance first locks the customer's row, so two
  checkouts at the same moment cannot both use the same available credit.

### Payment status of orders

| Order | `payment_method` | `payment_status` |
|---|---|---|
| cash, placed | `cash_on_delivery` | `pending` |
| cash, delivered and paid | `cash_on_delivery` | `paid` |
| cash, delivered, **not** paid (`payment_collected: false`) | `cash_on_delivery` | `deferred` (now in the debt) |
| pay later, placed / delivered | `pay_later` | `deferred` |

Debt is tracked per customer, not per order: a payment reduces the customer's
balance, not a specific order's status.

### Endpoints

| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/api/debt` | User | My account + ledger: `{account, transactions: page}` |
| GET | `/api/admin/debts?outstanding=true` | Admin | Customers with a credit limit or history, biggest balance first; `outstanding=true` = only those who owe |
| GET | `/api/admin/customers/{id}/debt` | Admin | One customer's account + ledger |
| PUT | `/api/admin/customers/{id}/credit-limit` | Admin | `{credit_limit}` (0 turns pay later off) |
| POST | `/api/admin/customers/{id}/debt/payments` | Admin | `{amount, note?, order_id?}`; `amount` is positive |
| POST | `/api/admin/customers/{id}/debt/adjustments` | Admin | `{amount, note, order_id?}`; `amount` is signed |

Account object:

```json
{ "user_id": 9, "first_name": "Sara", "last_name": "K", "email": "sara@example.com", "phone": "079...",
  "credit_limit": 250, "balance": 40, "available_credit": 210, "pay_later_allowed": true,
  "last_transaction_at": "2026-11-22T18:00:00+03:00" }
```

Ledger row:

```json
{ "id": 31, "user_id": 9, "order_id": 120, "type": "payment", "amount": -60,
  "balance_after": 40, "note": "cash at the shop", "created_by": 1, "created_at": "..." }
```

`created_by` is the admin who recorded it; `null` means automatic (checkout or cancellation).

Checkout with pay later:

```json
POST /api/orders
{ "address_id": 3, "payment_method": "pay_later" }
```

Delivering a cash order without collecting:

```json
PUT /api/admin/orders/120/status
{ "status": "delivered", "payment_collected": false }
```

### Errors

| Situation | Response |
|---|---|
| `payment_method` not `cash_on_delivery`/`pay_later` | 400 |
| pay later with credit limit 0 | 409 "pay later is not available for your account" |
| order total above available credit | 409 "this order (300.00) is more than your available credit (50.00)" |
| payment ≤ 0, adjustment = 0, adjustment without note, negative credit limit | 400 |
| payment above balance, or no debt at all | 409 |
| adjustment making the balance negative | 409 |
| `order_id` that is not this customer's | 404 |
| `payment_collected` sent when not delivering a cash order | 400 |

### Database

```
users                       + credit_limit numeric(10,2) DEFAULT 0, CHECK ≥ 0
customer_debt_transactions  id, user_id → users (RESTRICT), order_id → orders (SET NULL),
                            type, amount, note, created_by → users (SET NULL), created_at
orders                      payment_method ∈ {cash_on_delivery, pay_later}
                            payment_status ∈ {pending, paid, failed, deferred}
```

The table's `CHECK` constraints enforce the sign per type and the required
note on adjustments.

---

## 5. How the three features meet at checkout

`POST /api/orders` runs as **one transaction**:

1. Load the delivery address (must be yours; the default if none is given).
2. **Pay later:** lock your user row, so credit checks cannot race.
3. Lock your cart.
4. Lock **every** product the cart needs - single items and products inside
   bundles - in one statement, ordered by id (prevents deadlocks).
5. Check every line: products (active, stock, color) and bundles (rule in §3),
   then the **combined** demand per product. All problems are reported together.
6. Prices: products use `pricing.BestPrice` (own discount vs. live offers);
   bundles use their own price.
7. **Pay later:** `balance + total ≤ credit_limit`.
8. Insert the order, its items (with offer snapshot), its bundles (with their
   products), reduce stock once per product, and add a `charge` to the debt
   ledger if paying later.
9. Empty the cart (items and bundles).

If any step fails, nothing is saved: no order, no stock change, no debt.

---

## 6. Changes to existing endpoints

| Endpoint | Change |
|---|---|
| `GET /api/products`, `/api/products/{id}`, `/api/favorites` | `final_price` now includes live offers; new `offer` field |
| `GET /api/products` | new `?on_offer=true`; `sort=price_*` uses the final price |
| `GET /api/cart` and every cart endpoint | new `bundles` list; items have `offer`; `subtotal`/`item_count` include bundles |
| `DELETE /api/cart` | also removes bundles |
| `POST /api/orders` | new optional `payment_method` (`cash_on_delivery` default, `pay_later`) |
| order objects | new `bundles` list; items have `offer_id`, `offer_name`; `payment_method` can be `pay_later`, `payment_status` can be `deferred` |
| `PUT /api/admin/orders/{id}/status` | new optional `payment_collected` (only when delivering a cash order) |
| `GET /api/admin/orders` | new `?payment_method=` filter |
| order cancel | restores bundle stock too; reverses any debt charged for the order |
| reviews | products bought inside a bundle count as verified purchases |

All existing request bodies still work unchanged.

---

## 7. Files added and changed

**New**

| File | Purpose |
|---|---|
| `migrations/000002_offers_bundles_debts.up.sql` / `.down.sql` | schema for the three features |
| `internal/migrate/migrate.go` | applies migrations and tracks them in `schema_migrations` |
| `cmd/migrate/main.go` | `go run ./cmd/migrate up|status` |
| `internal/models/offer.go`, `bundle.go`, `debt.go` | data shapes and constants |
| `internal/validators/features.go` | input rules for offers, bundles, payments, adjustments, credit limits |
| `internal/repository/offer_repository.go` | `offerAppliesSQL` (the one SQL definition of "offer applies"), `loadLiveOffers`, `attachOffers`, offer CRUD |
| `internal/repository/bundle_repository.go` | bundle CRUD, live components, `bundleProblem` (the availability rule) |
| `internal/repository/cart_bundle_repository.go` | bundles in the cart |
| `internal/repository/debt_repository.go` | ledger helpers, accounts, statements, payments, adjustments |
| `internal/services/offer_service.go`, `bundle_service.go`, `debt_service.go` | business rules |
| `internal/handlers/offer_handler.go`, `bundle_handler.go`, `debt_handler.go` | HTTP endpoints |
| `tests/features_test.go`, `tests/migrate_test.go` | end-to-end tests |

**Changed**

| File | Change |
|---|---|
| `internal/pricing/pricing.go` | `OfferPrice`, `BestPrice` |
| `internal/models/product.go`, `cart.go`, `order.go` | offer / bundle / payment fields |
| `internal/repository/product_repository.go` | offers on lists and details, `on_offer` and offer filters, offer-aware price sort |
| `internal/repository/favorite_repository.go` | offers on favorites |
| `internal/repository/cart_repository.go` | offer-aware line prices, bundles in the cart view, clear removes bundles |
| `internal/repository/order_repository.go` | new checkout (bundles, combined stock, offers, pay later), bundle loading, cancel/deliver side effects |
| `internal/repository/review_repository.go` | verified purchase includes bundles |
| `internal/services/order_service.go`, `cart_service.go` | payment method, payment collected, cart bundles |
| `internal/handlers/order_handler.go`, `cart_handler.go`, `product_handler.go` | new fields, endpoints and filters |
| `internal/routes/routes.go` | 23 new routes |
| `tests/api_test.go` | test database is built with the migration tool |

---

## 8. Tests

```powershell
go test ./...                                   # unit tests
$env:INTEGRATION_DATABASE_URL = "postgres://USER:PASS@localhost:5432/postgres?sslmode=disable"
go test ./tests -v                              # end-to-end on a throwaway database
```

- **Unit:** `OfferPrice` / `BestPrice` (best single discount, no stacking, ties,
  fixed offers never below zero), offer/bundle/payment validation.
- **`TestOffers`:** validation, live vs. scheduled offers, best-price rule,
  category offers, public listing, `on_offer`, offer-aware sorting, the cart
  and order keep the offer, pausing and deleting.
- **`TestBundles`:** color required, computed totals and available stock,
  adding to the cart beyond stock, **combined stock** of a bundle plus single
  items, order snapshot, stock restored on cancel, verified reviews via a
  bundle, deactivation.
- **`TestCustomerDebts`:** pay later refused without a limit, credit limit,
  charge on checkout, over-limit refusal, reversal on cancel, payments
  (limits, running balance, admin recorded), adjustments, pay-later delivery
  stays deferred, cash not collected becomes debt, admin overview,
  permissions.
- **`TestMigrateAdoptsExistingDatabase`:** a database made the old way gets
  only `000002`, and a second run does nothing.

---

## 9. Limits and possible next steps

- **Coupon codes** (codes typed at checkout, usage limits) are not included;
  offers are automatic only.
- Offers apply to single products only, not to bundles.
- Debt is tracked per customer; payments are not allocated to specific orders,
  so an order's `payment_status` stays `deferred` after the customer pays.
  Per-order settlement could be added on top of the same ledger.
- No reminders or due dates for debts (needs a notification provider and a
  `due_date` column).
- There is no "credit" balance in the customer's favour: payments above the
  debt are refused.
