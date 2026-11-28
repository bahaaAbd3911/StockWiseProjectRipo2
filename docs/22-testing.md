# 22. Testing

## 22.1 How to run the tests

```powershell
cd backend
go vet ./...
go test ./...                                    # unit tests (no database needed)

$env:INTEGRATION_DATABASE_URL = "postgres://USER:PASS@localhost:5432/postgres?sslmode=disable"
go test ./tests -v                               # end-to-end tests

cd ..\Front-End
npm run lint
npm run build
```

**The end-to-end tests never touch your data.** They:

1. connect to the maintenance database you give;
2. create `ecommerce_it_<timestamp>`;
3. apply every migration with the same engine as production;
4. run the real router **as the restricted role `stockwise_app` under
   row-level security**, as in production;
5. drop the database at the end.

Emails go to an in-memory mailer, and uploads to a temporary folder.

**Race detector.** `go test -race` needs a C compiler. Without one locally,
run it in Docker against a throwaway Postgres:

```bash
docker network create sw-race
docker run -d --rm --name sw-race-db --network sw-race -e POSTGRES_PASSWORD=racetest postgres:18-alpine
docker run --rm --network sw-race \
  -e INTEGRATION_DATABASE_URL="postgres://postgres:racetest@sw-race-db:5432/postgres?sslmode=disable" \
  -v "$PWD:/src" -w /src golang:1.27 go test -race ./...
docker stop sw-race-db && docker network rm sw-race
```

## 22.2 Unit tests

| Package | Covers |
|---|---|
| `config` | defaults, every production refusal and warning, origins, TLS settings |
| `auth` | tokens, versions, store and platform issuers |
| `middleware` | refusing old token versions, store detection, client IP only behind a trusted proxy |
| `tenant` | host → store, caching, served hosts |
| `pricing` | best price, no stacking, ties, fixed never below 0, coupon discounts |
| `validators`, `models` | input rules, status transitions |
| `ratelimit` | burst, refill, independent keys, failure counting, clean-up |
| `images` | GPS removed, orientation kept, PNG/WebP metadata removed, fake / huge / broken images refused |
| `storage` | per-store keys, ownership checks |
| `notify` | retry delays, email contents, links per store |
| `server`, `httpx` | HTTP → HTTPS redirect, security headers, HSTS only on HTTPS in production, JSON limits |

## 22.3 End-to-end tests (`backend/tests`)

| Test | Proves |
|---|---|
| `TestFullShoppingFlow` | register → catalog → cart → checkout → statuses → reviews → waiting list, with permissions |
| `TestLastUnitCanOnlyBeSoldOnce` | two simultaneous checkouts for the last unit: exactly one wins |
| `TestAccountSecurity`, `TestLoginRevocation`, `TestPasswordReset`, `TestRateLimits` | account protection |
| `TestEmailVerification` | blocked checkout, resend, bad links, then checkout works |
| `TestOffers`, `TestBundles`, `TestCustomerDebts`, `TestCoupons` | promotions and pay later, including the coupon last-use race |
| `TestOrderAdminNotes`, `TestCSVExports` | notes saved, trimmed and limited, never shown to customers; CSV files with the byte-order mark, header, items, filters, Arabic text, formulas neutralised, admins only |
| `TestShippingFees` | settings validation, the cart estimate for the default and a chosen city, free over the threshold, the fee charged and kept on the order, the email line |
| `TestSalesScreens` | **the exact bodies the React sales screens send and every field they read**: coupon / offer / bundle forms, field errors, offers page, bundles in the cart, coupon totals, My Balance, pay-later checkout, statements, payments, corrections, Notify me, waiting lists |
| `TestOrderNotifications`, `TestEmailDeliveryIsRetried`, `TestDebtNotificationsAndReminders`, `TestBackInStockEmails`, `TestAdminEmailQueue` | emails and the queue |
| `TestImageUploads`, `TestUploadsArePerStore` | uploads and per-store folders |
| `TestProductWithImagesAndColors`, `TestCategories`, `TestStoreSettings`, `TestAdminUsers` | the shapes the React admin uses, and the user-management safety rules |
| `TestAdminDashboard` | local days, live figures, limits |
| `TestStoreFromHostName`, `TestStoresAreIsolated`, `TestDatabaseEnforcesIsolation` | the multi-store guarantees, down to raw SQL on the API's own connection |
| `TestPlatformAPIOnlyOnPlatformHost`, `TestPlatformCreatesReadyStore` | the platform admin |
| `TestStoreContentCustomization`, `TestCustomDomains` | store content and domains |
| `TestMigrateAdoptsExistingDatabase`, `TestCORSPreflightAndHealth` | tooling |

## 22.4 Continuous integration (GitHub Actions)

`.github/workflows/ci.yml` runs on every push to `main`, on every pull
request, and every Monday at 05:00 UTC (08:00 in Amman), so a newly
published vulnerability shows up even in a week without a push. It can
also be started by hand (Actions → CI → Run workflow). The result shows as a green tick or red cross next to
the commit on GitHub.

| Job | What it checks |
|---|---|
| Backend (Go) | `gofmt`, `go vet`, every Go test under the race detector, against a throwaway PostgreSQL 18 service |
| Frontend (React) | `npm ci` (exact locked versions), ESLint, the production build, all on Linux, so an import whose letter case is wrong fails here and not on the server |
| Go vulnerability scan and static analysis | `govulncheck`: known vulnerabilities in the Go code's dependencies and standard library; `staticcheck`: likely bugs and dead code |
| Docker images | both production images build, and the Caddyfile is valid (runs after the first two pass) |

**Why:** a mistake is caught minutes after it is pushed, not when a
customer finds it. `main` should only receive commits with a green tick.
On GitHub you can make this required: Settings → Branches → a rule for
`main` → "Require status checks to pass".

### Dependabot (`.github/dependabot.yml`)

Every Monday GitHub looks for newer versions of everything the project uses:

- Go modules;
- npm packages;
- GitHub actions;
- Docker base images: golang, distroless, node, caddy.

For each update it opens a **pull request**, and CI runs on it:

- **Green:** merge it on GitHub.
- **Red:** the update breaks something. Leave it open and ask for a fix.

Minor and patch updates come grouped, one pull request per area; major
versions come one by one, because they are the ones that can break things.
For security fixes as soon as they are published, turn on **Dependabot
security updates** (GitHub → Settings → Code security). This is the only
case where something other than you puts commits on `main`: by merging a
reviewed pull request.

## 22.5 The latest results

- `go vet`: clean.
- All unit and end-to-end tests pass, five runs in a row.
- All tests also pass under the race detector, on a fresh Postgres 18.
- `npm run lint` is clean and `npm run build` succeeds.

What testing found and fixed is listed in
[23-security-and-quality-review.md](23-security-and-quality-review.md).

## 22.6 Manual API checklist (Postman or similar)

Most of this is automated. Use it when clicking through the API by hand, on
a test database, never real data.

- **Setup:**
  - [ ] `GET /api/health` → `{"database":"ok","status":"ok"}`
  - [ ] register two customers and make a third an admin
- **Account:**
  - [ ] the same email in other capitals → 409
  - [ ] `ALICE@...` logs in as `alice@...`
  - [ ] wrong password → 401
  - [ ] wrong current password on change → 400 `fields.current_password`
  - [ ] after `DELETE /api/profile` the token gets 401
- **Permissions:**
  - [ ] no token → 401
  - [ ] customer token on `/api/admin/...` → 403
- **Catalog:**
  - [ ] bad product body → 400 listing every field
  - [ ] first image becomes primary
  - [ ] duplicate color → 409; `hex_code: "red"` → 400
  - [ ] filters and sorting work
  - [ ] a deleted product is hidden but still visible to admins
- **Cart and checkout:**
  - [ ] color missing → 400; more than the stock → 409 "only N left"; a second color → 409
  - [ ] no address → 400; someone else's address → 404
  - [ ] success → 201, stock reduced, cart empty
  - [ ] cancel → stock back; cancel again → 409
- **Admin orders:**
  - [ ] `pending → shipped` → 409; unknown status → 400
  - [ ] delivered sets `payment_status: paid`
- **Reviews, favorites, waiting list:**
  - [ ] a review after delivery is a verified purchase; a second review → 409
  - [ ] a helpful vote on your own review → 400; twice → 409
  - [ ] favorite twice → 201 then 200
  - [ ] waiting list for an in-stock product → 409

## 22.7 What is not automated

There are no browser tests: nothing clicks through the React screens
automatically. `TestSalesScreens` and the other frontend-shape tests check
the API side of every screen. Before a release, click through the
checklist in [21-deployment-and-operations.md](21-deployment-and-operations.md)
§21.4.
