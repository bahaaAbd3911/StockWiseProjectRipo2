# 9. Prices, offers, coupons and bundles

## 9.1 The pricing rule: one best price, never stacked

A product can have its own discount, and several offers can cover it at the
same time. The customer always gets the **single lowest price**:

```
candidates = [ price after the product's own discount,
               price after offer 1, price after offer 2, ... ]
final_price = the lowest candidate
offer       = the offer that produced it (null if the product's own discount won or tied)
```

How each candidate is computed:

- percentage: `price × (100 − value) / 100`, rounded to cents;
- fixed: `price − value`, never below 0.

Example (price 100, own discount 10%):

| Live offers | final_price | offer |
|---|---|---|
| none | 90.00 | null |
| a 25% sale | 75.00 | the 25% sale |
| a 25% sale and a 10.00-off offer | 75.00 | the 25% sale (not 65: no stacking) |
| only a 10.00-off offer (= 90, a tie) | 90.00 | null: the product's own discount wins a tie |

This rule lives in **one function**, `pricing.BestPrice`. It is used by:

- the product list and the product page;
- favorites;
- the cart;
- checkout.

The price a customer sees is therefore always the price they pay.

**Why no stacking?** Prices stay predictable for customers, and two
promotions cannot accidentally give a product away.

## 9.2 Offers

A timed price cut that applies **automatically**, with no code to type.

| Field | Meaning |
|---|---|
| `discount_type`, `discount_value` | `percentage` (0 < v ≤ 100) or `fixed` (amount off each unit) |
| `scope` | `products` (the ids in `product_ids`) or `category` (every product of one category, any letter case) |
| `starts_at`, `ends_at` | **required**; live while `starts_at ≤ now < ends_at` |
| `is_active` | pause without deleting |
| `status` | computed: `live`, `scheduled`, `expired`, `disabled` |

- Orders copy which offer set each item's price, and its name, so the order
  keeps it even if the offer is later deleted.
- Offers do not apply to bundles: a bundle's price is already the deal.
- `CHECK` constraints in the database repeat the main rules (type, value
  range, scope ⇔ category, window order), so bad data cannot be stored even
  by hand.

**The admin screen** (Offers) fills in "now" and "7 days from now" for a new
offer, because both dates are required. The customer screen (`/client/offers`)
lists live offers; choosing one shows its products, with prices that already
include the offer.

## 9.3 Coupon codes

A code the customer **types** in the cart, e.g. `SAVE10`.

| Setting | Meaning |
|---|---|
| `code` | 3–40 letters, digits, `-`, `_`; stored upper-case, typed in any case |
| `discount_type`, `discount_value` | `percentage` (0–100) or `fixed` amount |
| `scope` | `all` (default: products and bundles), `products` (listed ids), `category` (one category) |
| `min_order_amount` | cart subtotal needed (default 0) |
| `max_discount` | cap for percentage coupons, e.g. 20% but at most 25.00 |
| `starts_at`, `ends_at` | optional window |
| `usage_limit` | total uses (null = unlimited) |
| `per_customer_limit` | uses per customer: **once by default**; null = unlimited |
| `is_active` | on/off |

**Coupons stack on top of offers**: the coupon is taken off *after* each
line already has its best price. This was decided with the project owner. It
is simple to explain, and the store controls the cost with the minimum order
and the cap.

How the discount is worked out (`couponRow.evaluate` + `pricing.CouponDiscount`,
used by both the cart and checkout):

1. The coupon must be live (active, inside its window, uses left), and the
   customer must have uses left.
2. **Eligible amount** = the lines the scope covers, at their prices after
   offers. Bundles count only for `scope: all`, because they already have a
   special price.
3. The cart subtotal must reach `min_order_amount`. If it does not, the
   customer is told exactly how much to add.
4. Percentage: eligible × value / 100, rounded to cents, capped at
   `max_discount`. Fixed: the value. Never more than the eligible amount.

The flow:

- **Applying:** `POST /api/cart/coupon {code}` checks the code against the
  cart and attaches it, or answers 400 with the reason on `fields.code`. It
  has the strict per-IP rate limit, so codes cannot be brute-forced.
- **The cart shows** `coupon` (with `discount_amount`, or a `problem` if it
  stopped working: expired, used up, cart changed), plus `discount` and
  `total`.
- **Checkout re-checks the coupon with its row locked**, so two customers
  cannot both take the last use. If the coupon no longer works, checkout
  answers 409 instead of silently charging the full price.
- **The order keeps copies:** `subtotal_amount`, `discount_amount`,
  `coupon_code` and `total_amount` = subtotal − discount (a database check
  enforces it). Deleting the coupon later changes nothing.
- **Cancelling an order gives the use back.**

**The screens:** a coupon box in the cart summary; the admin Coupons page
lists each code with uses and total discount given, and edits all the
settings above.

## 9.4 Bundles

Several products sold together as **one cart line** at one price, e.g.
"Desk set = Keyboard (Black) + 2 × Mouse for 75.00".

| Field | Meaning |
|---|---|
| `price` | the price of one bundle |
| `items` | `[{product_id, quantity, selected_color}]`: at least 2 products, or 1 product with quantity ≥ 2 ("3 for 2"); at most 20; each product once |
| `selected_color` | required when the product has colors. The bundle fixes the color |
| `starts_at`, `ends_at` | optional window |
| `is_active` | on/off; `DELETE` deactivates |

Computed fields:

| Field | Meaning |
|---|---|
| `original_total` | what the items cost bought separately today (at their best prices) |
| `savings` | `original_total − price`, never below 0 |
| `available_stock` | how many complete bundles the stock allows: `min(stock ÷ quantity)` |
| `is_available`, `unavailable_reason` | whether one can be bought now, and why not |

**The availability rule** (`bundleProblem`, used by the bundle pages, cart
and checkout). A bundle cannot be bought when:

1. it is not live;
2. a product inside it was deactivated;
3. its color for a product was deleted;
4. a product inside it gained colors after the bundle was made;
5. there is not enough stock for `bundle quantity × item quantity`.

**Combined stock.** A product can be in the cart on its own *and* inside a
bundle. Each line can look fine alone while together they need more than
the stock. Checkout adds everything up per product and refuses with, e.g.,
"Mouse: only 4 left in stock, but your cart needs 5".

**Why a separate cart table for bundles?** `cart_items` requires a product
and allows one row per product; a bundle is neither.

Orders keep a snapshot of each bundle: name, price, quantity, and its
products with their colors. Effects later on:

- **cancelling** gives back the stock of every product inside;
- **delivering** marks reviews of those products as verified purchases.

**The screens:** the customer Bundles page (`/client/bundles`) shows
contents, price, "bought separately" and savings; bundle lines in the cart
and checkout have their own quantity buttons. On the admin Bundles page an
admin picks products, quantities and colors, and the page shows what the
products would cost separately, to help choose a price.
