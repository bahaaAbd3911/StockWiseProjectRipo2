# 10. Cart, checkout and orders

## 10.1 The cart

- **Product lines.** There is one line per product. Adding the same product
  in a second color is refused with a hint, so the line's quantity and color
  stay unambiguous.
- **Bundle lines** are a separate list.
- **Every cart response is live:**
  - each line: current price, line total, `is_available` and
    `unavailable_reason`;
  - for the whole cart: `subtotal`, the `coupon` (if any), `discount` and
    `total`.
- **Availability uses the same rule as checkout** (`cartLine.problem`). The
  cart never promises something checkout will refuse.

**The frontend uses the backend's totals.** The cart summary, the checkout
summary and the "Place order" button show `subtotal`, `discount` and `total`
exactly as the backend computed them. They are not re-added in the browser,
because offers, bundles and coupons are only known correctly on the server.

## 10.2 Checkout: one transaction

`POST /api/orders {address_id?, payment_method?}` does all of this, or
nothing:

1. **Address**: the given one (must be yours) or your default.
2. **Pay later**: lock your user row, so two checkouts cannot spend the same
   credit.
3. **Lock your cart**, so the same cart cannot be checked out twice.
4. **Lock every product** the cart needs, single items and products inside
   bundles, in one statement **ordered by id**. This prevents overselling,
   and the fixed order prevents deadlocks.
5. **Check** every line and the combined quantity per product. All problems
   are reported together in one 409.
6. **Prices**: `pricing.BestPrice` for products, the bundle price for
   bundles.
7. **Coupon**: re-checked with its row locked; the discount is computed.
8. **Pay later**: balance + total must fit the credit limit.
9. **Insert the order**, with copies of the address, names, prices, colors
   and offer. Reduce stock once per product, add a debt `charge` for pay
   later, and record the coupon use.
10. **Empty the cart and commit.** Then queue the "order received" email.

**Why one transaction with locks?** Without locks, two customers can both
read "1 left" and both buy it. The test `TestLastUnitCanOnlyBeSoldOnce` fires
two checkouts at the same instant for the last unit; exactly one succeeds.

Checkout needs a **confirmed email** ([07](07-accounts-and-security.md)).

## 10.3 Payment methods

| `payment_method` | Meaning |
|---|---|
| `cash_on_delivery` | default |
| `pay_later` | charged to the customer's debt account ([11](11-pay-later-and-debts.md)) |

On the checkout screen, "Pay Later" only appears for customers the store
gave a credit limit, and it is greyed out when the order is larger than the
available credit. The backend checks again either way.

## 10.4 The order life cycle

```
pending → confirmed → processing → shipped → delivered
   └──────────┴────────────┴──→ cancelled
```

`models.CanTransitionOrder` decides; any other move answers 409.

| Event | What happens |
|---|---|
| customer cancels | allowed while `pending` |
| admin cancels | allowed until `processing` |
| cancelled | stock given back (also for products inside bundles), debt charge reversed, coupon use given back |
| delivered, cash | `payment_status` becomes `paid`, or `deferred` + a debt charge when the admin sends `payment_collected: false` (the customer did not pay the driver) |
| delivered, pay later | stays `deferred` |
| delivered, any | the customer's reviews of those products become verified purchases |

The customer gets an email at every step ([13](13-emails.md)).

## 10.5 Why orders keep copies

An order stores the product names, prices, colors, the offer name, the
coupon code and the delivery address **as they were at checkout**. Renaming
a product, changing a price, deleting a coupon or editing an address later
never changes a past order. Invoices and history must not move.
