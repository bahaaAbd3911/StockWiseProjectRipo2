# StockWise documentation

This folder explains the whole project: what it does, how every part works,
how it was built, and **why** each decision was made. Every file covers one
topic and can be read on its own. Read them in order the first time.

| # | File | What it explains |
|---|---|---|
| 1 | [01-overview.md](01-overview.md) | what StockWise is, who uses it, the big picture |
| 2 | [02-history.md](02-history.md) | how the project was built, step by step, and why each step |
| 3 | [03-repository-layout.md](03-repository-layout.md) | every folder and what lives there |
| 4 | [04-request-lifecycle.md](04-request-lifecycle.md) | one click followed from the browser to the database and back |
| 5 | [05-multi-tenancy.md](05-multi-tenancy.md) | many stores in one system: how they are told apart and kept apart |
| 6 | [06-backend-architecture.md](06-backend-architecture.md) | the Go code: layers, packages, the rules every file follows |
| 7 | [07-accounts-and-security.md](07-accounts-and-security.md) | sign-up, login, email confirmation, password reset, revocation, rate limits |
| 8 | [08-catalog-and-uploads.md](08-catalog-and-uploads.md) | products, categories, colors, images and file uploads |
| 9 | [09-prices-offers-coupons-bundles.md](09-prices-offers-coupons-bundles.md) | how a price is decided; offers, coupon codes and bundles |
| 10 | [10-cart-checkout-orders.md](10-cart-checkout-orders.md) | the cart, the checkout transaction, the order life cycle |
| 11 | [11-pay-later-and-debts.md](11-pay-later-and-debts.md) | credit limits and the debt ledger |
| 12 | [12-reviews-favorites-addresses-waiting-list.md](12-reviews-favorites-addresses-waiting-list.md) | the smaller customer features |
| 13 | [13-emails.md](13-emails.md) | every email, the email queue and its worker |
| 14 | [14-admin-features.md](14-admin-features.md) | dashboard, user management, store look and content, custom domains |
| 15 | [15-platform-admin.md](15-platform-admin.md) | the people who run the platform and create stores |
| 16 | [16-frontend.md](16-frontend.md) | the React app: structure, data flow, every screen |
| 17 | [17-database-and-migrations.md](17-database-and-migrations.md) | the tables, the numbered migrations, how the code respects the schema |
| 18 | [18-configuration.md](18-configuration.md) | every setting and the production safety checks |
| 19 | [19-api-reference.md](19-api-reference.md) | every endpoint, who may call it, what it takes |
| 20 | [20-running-locally.md](20-running-locally.md) | running everything on your own computer |
| 21 | [21-deployment-and-operations.md](21-deployment-and-operations.md) | going live: Docker, Caddy, HTTPS, DNS, backups, monitoring |
| 22 | [22-testing.md](22-testing.md) | what is tested, how, and how to run the tests |
| 23 | [23-security-and-quality-review.md](23-security-and-quality-review.md) | the full security and optimization checkup and what it found |
| 24 | [24-decisions.md](24-decisions.md) | every important choice, the alternatives, and why |
| 25 | [25-limits-and-next-steps.md](25-limits-and-next-steps.md) | what is not done yet and what to do next |
| 26 | [26-glossary.md](26-glossary.md) | the words used in these documents |

**If you only have ten minutes**, read 01, 05 and 24.
**If you are about to deploy**, read 18 and 21.
**If you are changing code**, read 04, 06 and the file of the feature you touch.
