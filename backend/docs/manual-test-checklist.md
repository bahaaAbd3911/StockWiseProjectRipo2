# Manual API checklist (Postman)

Most of this is automated in `tests/api_test.go`. Use the list below when
clicking through the API by hand. Use a test database, not real data.

## Setup
- [ ] `GET /api/health` → `{"database":"ok","status":"ok"}`
- [ ] Register two customers; promote a third account with
      `UPDATE users SET role='admin' WHERE email='...'`
- [ ] Log in; keep each `token` in a Postman variable

## Account
- [ ] Registering the same email in different capitals → 409
- [ ] Logging in with `ALICE@...` works for `alice@...`
- [ ] Wrong password → 401 `invalid email or password`
- [ ] `PUT /api/profile/password` with the wrong current password → 400, `fields.current_password`
- [ ] After `DELETE /api/profile` the same token gets 401

## Permissions
- [ ] No token on a protected route → 401
- [ ] Customer token on `/api/admin/...` or `POST /api/products` → 403

## Catalog
- [ ] Create a product with a bad body → 400 listing every bad field
- [ ] First image becomes primary; adding another primary moves the flag
- [ ] Duplicate color name → 409; `hex_code: "red"` → 400
- [ ] `GET /api/products?category=...&search=...&sort=price_asc` filters and sorts
- [ ] `DELETE /api/products/{id}` hides it from `/api/products` but `/api/admin/products/{id}` still shows it

## Cart & checkout
- [ ] Product with colors and no `selected_color` → 400
- [ ] More than the stock → 409 with "only N left"
- [ ] Same product in another color → 409
- [ ] Cart `subtotal` uses the discounted price
- [ ] Checkout without any address → 400; with someone else's `address_id` → 404
- [ ] Successful checkout: 201, stock reduced, cart empty
- [ ] Cancel while pending → stock restored; cancel again → 409

## Admin orders
- [ ] `pending → shipped` → 409; `"status":"lost"` → 400
- [ ] `confirmed → processing → shipped → delivered` works; delivered sets `payment_status: paid`

## Reviews, favorites, waiting list
- [ ] Review after delivery → `verified_purchase: true`; second review → 409
- [ ] Helpful vote on your own review → 400; twice → 409; removing it lowers the count
- [ ] Favorite twice → 201 then 200
- [ ] Waiting list for an in-stock product → 409; duplicate → 409
- [ ] Admin `notified` sets `notified_at`; a cancelled request cannot be changed
