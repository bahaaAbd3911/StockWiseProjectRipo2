# 23. Security and quality review

A full checkup of the code for safety and optimization, done after the
sales screens were added. This file lists what was checked, how, what was
found, and what was changed.

## 23.1 Automated checks

| Check | Result |
|---|---|
| `go vet ./...` | clean |
| all backend unit and end-to-end tests | pass |
| the same tests under Go's **race detector** (unit tests, and end-to-end on a fresh Postgres 18 in Docker) | pass, no data races |
| end-to-end tests run five times in a row | pass every time (after the fix in 23.3) |
| `npm run lint` | clean |
| `npm run build` | succeeds |
| `npm audit` | 3 moderate findings, all in `minimatch` inside ESLint: a **development tool only**, not part of what users download. Fix with `npm audit fix` when convenient |
| `govulncheck` | now runs in CI on every push. Its first run found **GO-2026-6859** (an HTTP/1 server crash via a `MaxBytesReader` race in Go's `net/http`, which the API uses to cap request sizes). Go 1.27.0–1.27.3 are affected; fixed by moving to **Go 1.27.4** (`go.mod`, `backend/Dockerfile`) |
| `staticcheck` | **not run**: downloading and running it was blocked in this environment. Run it once yourself: `go run honnef.co/go/tools/cmd/staticcheck@latest ./...` |

## 23.2 Manual review: what was checked and found sound

| Area | What was checked | Finding |
|---|---|---|
| SQL injection | every place where SQL text is joined | only fixed fragments written in the code; all values use `$N` placeholders; sort orders come from a whitelist |
| Login tokens | parsing options | HS256 pinned (no `alg: none`), issuer and expiry required, version checked against the database, store bound |
| Store isolation | host-only store choice, RLS, foreign keys, tests | sound; proven down to raw SQL on the API's connection |
| Request size | JSON and uploads | JSON bodies are size-capped (`MaxBytesReader`); uploads are capped and checked by content |
| Slow clients | server timeouts | read-header 5 s, read 60 s, write 30 s, idle 60 s |
| Huge lists | pagination | `limit` is capped at 100 |
| Brute force | rate limits | per IP, per email, per store; coupon codes rate-limited |
| Error leaks | error writer | unknown errors return only "internal server error" |
| Frontend XSS | `dangerouslySetInnerHTML`, `eval`, `innerHTML` | none used; React escapes all text |
| Secrets in git | `.env` files | ignored; only examples are committed |

## 23.3 Problems found and fixed

| # | Problem | Risk | Fix |
|---|---|---|---|
| 1 | **The React pages had no security headers.** The API sent strict ones, but the pages Caddy serves had none | another site could frame the admin panel (clickjacking); any injected script could read the login token from localStorage | the Caddyfile now sends a Content-Security-Policy (only the app's own scripts; fonts and icons from Google Fonts and cdnjs; https images), `X-Frame-Options: DENY`, `nosniff`, a referrer policy and HSTS on every page. Validated with `caddy validate`; the build has no inline scripts, so the policy does not block the app |
| 2 | **Links in order, balance and back-in-stock emails were broken.** Emails link to `/orders/{id}`, `/account/debt` and `/products/{id}`, which the app did not have, so customers landed on the home page | customers could not reach their order or balance from an email | new forwarding routes (`Pages/account/EmailLinks.jsx`) open the My Orders tab, the My Balance tab, or the product page |
| 3 | **Offers form sent no dates** (found by the new `TestSalesScreens` before release) | creating an offer from the admin screen always failed | the dates are required and default to now → 7 days |
| 4 | **A flaky test**: `TestOrderNotifications` failed about one run in three | false alarms hide real failures | the test checked the database before the worker had recorded `sent_at`; it now waits up to 5 s, like the rest of the test |
| 5 | **`useToast` returned a new object every render**; any effect depending on it would run in a loop | an endless request loop on pages that reload after an action | `useToast` is now memoized; `openProfileTab` too |
| 6 | **Uncacheable assets / stale pages**: Caddy set no cache rules | slower repeat visits, or an old `index.html` after a release | `/assets/*` (hashed names) cached for a year; pages `no-cache` |
| 7 | **Product cards and pages ignored offers.** They worked out the price from the product's own discount only, although the backend already sends the best price | shoppers saw a higher price than the cart charged, and the Offers page showed prices without the offer | `finalPrice` now uses the backend's price (offer included); the struck-through price, the "-X%" badge and "Save X%" use the real saving, and the product page names the offer |

## 23.4 Optimization

| Finding | Change | Result |
|---|---|---|
| Shoppers downloaded the whole admin panel and platform console in one 824 KB JavaScript file | admin pages, the admin layout and the platform console load with `React.lazy` only when opened | the shop's main file is **582 KB** (182 KB gzipped), −29% |
| The build contained **9.0 MB of images**, e.g. `perfume.jpg` 2.4 MB (5184 px wide), two 1.4 MB illustrations shown at 500 px, two 1.3 MB backgrounds, and a 461 KB favicon | photos resized to 1920 px and recompressed as JPEG; illustrations and backgrounds converted to WebP at twice their display size (transparency kept); icons resized to 160 px; favicon replaced by a 64 px PNG (`public/favicon.png`). Compared side by side: no visible difference. Originals are in git history | images **9.0 MB → 0.55 MB**; the whole build **10.7 MB → 1.7 MB** |
| Dashboard and list queries | indexes on `orders.created_at` and `status` exist (000005); the dashboard reads one snapshot | no change needed |
| Store lookups on every request | host → store is cached for 30 s; the store setting is sent to the database only when it changes | no change needed |

## 23.5 Known trade-offs (accepted, documented)

- **Login tokens are kept in localStorage**, which scripts on the page can
  read. The CSP above is what protects it. Moving to http-only cookies would
  need CSRF protection and a different login flow.
- **Rate-limit counters are in memory**: fine for one API server; use Redis
  with more.
- **Failed-login lock**: someone who knows a customer's email can lock their
  login for 15 minutes; the customer can still reset the password.
