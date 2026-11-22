# Backend architecture - how everything fits together

## 1. The big idea: layers

Every request travels through the same layers, and each layer has one job:

```
HTTP request
  │
  ▼  middleware/   Recover → Logger → CORS        (every request)
  ▼  routes/       which handler? which access level?
  ▼  middleware/   Authenticate → RequireAdmin    (protected routes only)
  ▼  handlers/     read URL + JSON, call ONE service, write JSON
  ▼  services/     validate, apply business rules, combine repositories
  ▼  repository/   SQL, transactions, row locks
  ▼  PostgreSQL    constraints are the last safety net
```

Rules that keep it clean:

| Layer | May know about | Must NOT contain |
|---|---|---|
| handlers | HTTP, JSON, services | SQL, business rules |
| services | models, validators, repositories | HTTP status codes, SQL |
| repository | SQL, models | HTTP |

Errors flow back up the same way. A repository or service returns a typed
error (`apperr.NotFound("order not found")`). The handler just returns it, and
`httpx.WriteError` turns the error type into a status code. **No code
anywhere compares error strings.**

## 2. Folder by folder

### `cmd/api/main.go` - the program entry point
Loads `.env` → `config.Load()` → `database.Connect()` → `routes.New()` →
starts the HTTP server with timeouts → waits for Ctrl+C/SIGTERM → graceful
shutdown (running requests get 10 s to finish).

### `internal/config/` - settings
`config.go` reads `APP_PORT`, `DATABASE_URL`, `JWT_SECRET`, `JWT_TTL` and
`CORS_ALLOWED_ORIGINS` once, applies defaults, and **refuses to start** if the
database URL is missing or the JWT secret is shorter than 32 characters. No
other package reads environment variables.

### `internal/database/` - connection pool
`postgres.go`
- `Connect` opens a pool (max 25 connections) and pings the database.
- `WithTx(ctx, db, fn)` runs `fn` in a transaction: commit if it returns nil,
  roll back otherwise. Every multi-step write uses it.
- `Querier` is an interface satisfied by both `*sql.DB` and `*sql.Tx`, so a
  helper can run inside or outside a transaction.

### `internal/apperr/` - the error vocabulary
`apperr.go` defines five kinds of expected errors: `Invalid` (400),
`Unauthorized` (401), `Forbidden` (403), `NotFound` (404), `Conflict` (409).
`Validation(fields)` is an `Invalid` error with per-field messages.
`FromDB` translates database errors: `sql.ErrNoRows` → NotFound, unique
violation → Conflict, foreign-key violation → NotFound, check/length → Invalid.
Anything else stays an unknown error and becomes a 500.

### `internal/httpx/` - HTTP helpers
`httpx.go`
- `Handle` adapts our `func(w, r) error` handlers to `http.Handler` and sends
  any returned error to `WriteError`.
- `WriteError` maps error kinds to status codes. Unknown errors are **logged**
  and the client only sees `internal server error`, so SQL never leaks.
- `DecodeJSON` limits bodies to 1 MiB and gives clear messages for an empty
  body, bad JSON or a wrong field type.
- `JSON`, `Message`, `PathID`, `QueryInt`, `Pagination`.

### `internal/auth/` - tokens and identity
`auth.go`
- `TokenManager.Issue(userID)` signs an HS256 JWT with `user_id`, issuer and
  expiry. `Parse` checks the signature, the algorithm (blocks the `alg: none`
  trick), the issuer and the expiry.
- `Identity{UserID, Role}` is stored in the request context by the middleware.
  Handlers call `auth.UserID(r.Context())`.

### `internal/middleware/`
- `auth.go`
  - `Authenticate` checks the Bearer token, then asks the database for the
    user's **current** role and `is_active`. A deactivated account is locked
    out immediately, and the role can never come from the client.
  - `RequireAdmin` returns 403 unless the role is `admin`.
- `cors.go` lets the frontend origin(s) call the API and answers browser
  preflight requests.
- `logging.go`: `Logger` writes one line per request; `Recover` turns a panic
  into a 500 instead of a dropped connection.

### `internal/routes/routes.go` - the map of the API
Creates every repository → service → handler (dependency injection), then
registers each URL with one of three helpers: `public(...)`, `user(...)`
(needs login) or `admin(...)` (needs login and the admin role). Reading this one
file shows every endpoint and who may call it.

### `internal/models/` - data shapes and pure domain rules
| File | Contents |
|---|---|
| `common.go` | `Page[T]` for paginated lists. Makes money serialise as JSON numbers |
| `user.go` | `User` (has no password field, so a hash can never be sent), register/login/profile requests |
| `product.go` | `Product` (+ computed `final_price`, `in_stock`, `primary_image_url`, `rating_average`, `review_count`), `ProductDetail`, images, colors, filters |
| `cart.go` | `Cart`/`CartItem` with live prices and availability |
| `favorite.go` | `Favorite` embedding the product |
| `address.go` | `Address`, `AddressInput` |
| `order.go` | `Order`, `OrderItem`, statuses, and **`CanTransitionOrder`**, the order life-cycle rule |
| `review.go` | `Review` (reviewer shown as "First L."), `ReviewInput` |
| `waiting_list.go` | `WaitingListItem`, statuses, `CanTransitionWaiting` |

Money is `decimal.Decimal` everywhere, never `float64`, to match
`numeric(10,2)` exactly.

### `internal/pricing/pricing.go` - the money math
`DiscountedPrice(price, percent)` and `LineTotal(unit, qty)`. This is the
**only** place a discount is applied, so the product page, cart and checkout
always show the same number.

### `internal/validators/` - input checks
- `validator.go`: the `Checker` collects every field problem at once; there are
  helpers for required fields, character-length limits (counted in characters,
  like `varchar`), money (0..99,999,999.99, 2 decimals), percent, phone, email,
  URL and hex colors.
- `rules.go`: one `Validate...` function per request. Each one first
  **normalises** the input (trims spaces, lower-cases emails, turns `""` into
  NULL) and then checks it against the schema limits.

### `internal/services/` - business rules
| File | What it decides |
|---|---|
| `user_service.go` | Register (unique email, bcrypt hash), login (same response time for unknown emails, generic error), change password, deactivate |
| `product_service.go` | List/sort/filter, product detail = product + images + colors, admin create/update/deactivate, images and colors |
| `cart_service.go` | Default quantity 1, validation; every change returns the updated cart |
| `favorite_service.go` | Add (idempotent), list, remove |
| `address_service.go` | Validation, then the repository |
| `order_service.go` | Checkout, customer cancel, admin status changes (status must be valid; nothing goes back to `pending`) |
| `review_service.go` | Product must be active; `verified_purchase` is decided by the server |
| `waiting_list_service.go` | Join, cancel, admin status |

### `internal/repository/` - SQL
| File | Highlights |
|---|---|
| `shared.go` | `productColumns` + `scanProduct` (shared by products and favorites), `resolveColor` (shared color rule for cart and waiting list), `whereBuilder` for safe dynamic filters |
| `user_repository.go` | Case-insensitive email lookup, `AuthInfo` for the middleware |
| `product_repository.go` | Paginated list with whitelisted sorts, soft delete, categories |
| `catalog_repository.go` | Images (first is primary, one primary, promote on delete), colors |
| `cart_repository.go` | `cartLine.problem()` - **the one availability rule** used by both the cart view and checkout |
| `favorite_repository.go` | `ON CONFLICT DO NOTHING` |
| `address_repository.go` | User-row lock for the one-default rule, auto-default, promote on delete |
| `order_repository.go` | **Checkout transaction**, items loaded in one query, cancel/status changes with stock restore |
| `review_repository.go` | One review per product, helpful vote and counter in one transaction |
| `waiting_list_repository.go` | Advisory lock against duplicates, status transitions, `notified_at` |

## 3. The important flows

### Checkout (`OrderRepository.Checkout`)
Everything runs in **one transaction**:
1. Load the address. It must belong to the user; if none is given, the default is used.
2. Lock the cart row, so the same user cannot check out twice at once.
3. Load the cart lines and **lock the product rows** in product-id order, so
   two customers cannot both buy the last unit and the locks cannot deadlock.
4. Check every line with `cartLine.problem()`: active? enough stock? does the
   color still exist? Report **all** problems in one 409 response.
5. Compute `unit_price = price × (100 − discount) / 100` (rounded to cents) and
   `subtotal = unit_price × qty`, exactly.
6. Insert the order with a **copy** of the address.
7. Insert the items with **copies** of name/price/color, and reduce stock.
8. Empty the cart and commit.

If anything fails, nothing is saved. The integration test
`TestLastUnitCanOnlyBeSoldOnce` fires two checkouts at the same time for a
product with stock 1 and checks that exactly one succeeds.

### Order life cycle
`models.CanTransitionOrder` is the single source of truth:
`pending → confirmed → processing → shipped → delivered`, plus cancellation
from the first three. Cancelling puts the stock back. Delivering sets
`payment_status = paid` and upgrades the customer's reviews of those products to
verified purchases.

### Authentication on every protected request
Token → signature/expiry check → database lookup (`is_active`, `role`) →
`Identity` in the context → handler uses `auth.UserID`. Every repository query
on personal data has `WHERE user_id = $n`, so asking for someone else's order or
address returns 404, as if it didn't exist.

## 4. How the code relies on the schema

| Schema rule | Where the code respects it |
|---|---|
| `varchar(n)` limits | validators check character counts first → 400 with a field name, not a DB error |
| `numeric(10,2)` money | `decimal` + `Money()` validator (≤ 2 decimals, ≤ 99,999,999.99) |
| `check_order_item_subtotal` | `LineTotal` is exact because `unit_price` is already rounded |
| `cart_items UNIQUE(cart_id, product_id)` | one line per product. Adding it in another color → 409 with a hint to change the line's color |
| `unique_default_address_per_user` | user-row lock + unset-then-set |
| `unique_primary_image_per_product` | product-row lock + unset-then-set |
| `unique_user_product_review`, `unique_review_user_vote` | friendly 409 messages |
| `orders.user_id ON DELETE RESTRICT` | accounts are deactivated, never deleted |
| `order_items.product_id ON DELETE SET NULL` | products are soft-deleted; order history keeps its copies |
| `check_payment_method = cash_on_delivery` | the only method the API uses |
| no unique index for active waiting-list entries | `pg_advisory_xact_lock` + an existence check |
| `users.created_at` nullable | `COALESCE` in the queries |

## 5. Tests

- Unit tests sit next to the code (`*_test.go`): pricing, validators, status
  rules, tokens, middleware, error mapping. Run them with `go test ./...`.
- `tests/api_test.go` runs the real router against a real, throwaway
  PostgreSQL database built from `migrations/000001_baseline.up.sql`. It covers
  registration through delivery, reviews, the waiting list, soft deletes,
  permissions, the last-unit race and CORS.
