# 14. Store admin features

Store admins work under `/admin` on their store's own address. The sidebar
holds these pages:

| Page | What it does | Details |
|---|---|---|
| Dashboard | the overview below | 14.1 |
| Products | create, edit, deactivate; images, colors, stock | [08](08-catalog-and-uploads.md) |
| Categories | names, icons, on/off | [08](08-catalog-and-uploads.md) |
| Orders | list, filter, change status, payment collected | [10](10-cart-checkout-orders.md) |
| Users | user management | 14.2 |
| Coupons | codes, limits, usage | [09](09-prices-offers-coupons-bundles.md) |
| Offers | timed price cuts | [09](09-prices-offers-coupons-bundles.md) |
| Bundles | product sets | [09](09-prices-offers-coupons-bundles.md) |
| Customer Debts | credit limits, payments, statements | [11](11-pay-later-and-debts.md) |
| Waiting List | who waits for what | [12](12-reviews-favorites-addresses-waiting-list.md) |
| Emails | whether emails go out; retry or cancel them | [13](13-emails.md) |
| Store Settings | the store's look and content, and shipping fees | 14.3, [10](10-cart-checkout-orders.md) |
| Domains | custom domains | 14.4 |

## 14.1 The dashboard

`GET /api/admin/dashboard?days=30` (1–366) returns everything in one call:

| Section | Contents |
|---|---|
| `today`, `period`, `previous_period` | orders, revenue, average order, items sold (bundle contents included), coupon discounts, cancelled orders, new customers; the previous period has the same length, for "+12% vs last month" |
| `sales_by_day` | one entry per day, days without orders included (ready for a chart) |
| `payment_methods` | orders and revenue for cash and pay later |
| `orders_by_status` | all orders right now |
| `best_sellers` | top 10 products by units (alone + inside bundles) |
| `low_stock` | active products at or below `LOW_STOCK_THRESHOLD` (5), with how many customers wait for each |
| `debts` | total owed, customers owing, customers over their limit, payments in the period |
| `waiting_list` | the most-requested out-of-stock products |
| `top_coupons` | the codes used most, with the discount given |
| `customers` | active, unconfirmed, new, ordering in the period |

**Revenue** means the value of orders placed (after coupons) and not
cancelled: booked sales, whether paid, owed or still to be collected.

How the numbers are kept honest:

- **Days follow the store's clock** (`STORE_TIMEZONE`, default
  `Asia/Amman`): "today" starts at midnight in Amman, not at midnight UTC.
- **All numbers come from one database snapshot** (a read-only REPEATABLE
  READ transaction), so they agree with each other even while orders arrive.

## 14.2 User management

Admins list and search users (with orders count and total spent), create,
edit and deactivate them, and open one user's page (orders and favorites).

The safety rules:

- **An admin cannot remove their own admin rights**, so nobody locks
  themselves out by mistake.
- **The last admin can never be removed or demoted**, so a store always
  has someone who can manage it.
- Deleting a user with orders deactivates them instead, because orders
  refer to them.
- A new password, a role change or a deactivation logs that user out
  everywhere.
- Accounts created by an admin count as email-confirmed.

## 14.3 Store look and content (store customization)

Each store has one settings row:

- **name**;
- **logo**: an uploaded image, a data URL up to 1 MB, or an http(s) URL;
- **theme colors** for light and dark mode;
- **page content**:
  - the home-page banner and the products-page banner (titles, texts and
    button in English and Arabic, plus an uploaded image);
  - up to 8 **feature tiles** (icon + texts);
  - the **footer**: tagline, copyright, social links.

**Empty fields keep the template's original texts**
(`src/data/storeContent.js`), so a new store looks complete from the first
second and an admin changes only what they want.

**One set of components renders every store** from these settings: nothing
in the code is specific to one store. The colors are applied as CSS
variables, so the whole theme changes without touching any component.

**Validation** stops a store from injecting anything into its pages:

- image URLs must be http(s) or uploaded;
- icons must be Font Awesome class names;
- social links must be one of the known networks;
- colors must be hex colors.

Replacing a banner image deletes the old file.

The same page has a **Shipping** card: the standard fee, cities with their
own fee (picked from the checkout's city list, so the names always match),
and "free shipping from". "Reset to default" resets the look only and keeps
the shipping fees. The rules are in [10-cart-checkout-orders.md](10-cart-checkout-orders.md) §10.3.

## 14.4 Custom domains

The Domains page:

1. Add a domain.
2. The page shows the two DNS records to create:
   - `TXT _stockwise-verify.<domain>` with the token;
   - `CNAME <domain> → <store>.<platform domain>`.

   For a root domain without `www`, use an ALIAS/ANAME record or the
   server's IP address instead of the CNAME.
3. Press **Verify**.
4. Once verified, make it the **main address**, which is used in emails.

How verification and certificates work, and why: [05-multi-tenancy.md](05-multi-tenancy.md).
