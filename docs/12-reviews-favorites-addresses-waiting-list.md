# 12. Reviews, favorites, addresses and the waiting list

## 12.1 Reviews

- One review per customer per product: a rating from 1 to 5 and a comment
  in English and/or Arabic.
- **"Verified purchase" is decided by the server**: true once an order
  containing that product, alone or inside a bundle, was delivered. A client
  cannot claim it.
- Helpful votes: once per review, and not on your own review.
- Sorting: newest, most helpful, rating high or low.
- Only the author can edit or delete a review.

## 12.2 Favorites

- Adding twice is harmless: 201 the first time, 200 "already saved" after.
  A double click never shows an error.
- Each favorite contains the product with its live price.

## 12.3 Addresses

- The first address becomes the default.
- **Exactly one default**, enforced by a unique index and a lock on the
  user's row.
- Deleting the default promotes the newest remaining address.
- Checkout uses the given address or the default. An address must belong to
  the customer, otherwise 404.

## 12.4 The waiting list

Customers can ask to be told when an out-of-stock product comes back.

- They can join **only for products that are out of stock**, with one
  active request per product (enforced with an advisory lock, so a double
  click cannot create two).
- Statuses: `waiting ⇄ notified → fulfilled / cancelled`.
- **The email is automatic.** When an admin saves the product with stock
  above 0, every `waiting` request becomes `notified` and the customer gets a
  "back in stock" email. Each request is told once; cancelled requests
  never.
- An admin can also mark one request `notified` by hand, which sends the
  email too.

**The screens:**

- **Product page:** a sold-out product shows **"Out of stock · Notify me"**
  instead of "Add to cart", using the chosen quantity and color.
- **Profile → Waiting List:** each request with its status ("Waiting",
  "Back in stock"…) and a Remove button.
- **Admin → Waiting List:**
  - who waits for what, filtered by status;
  - the current stock;
  - "Notify now" (sends the email) and "Cancel".
