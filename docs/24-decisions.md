# 24. Decisions and why

Every important choice, what else could have been done, and why this way
won.

## Architecture

| Decision | Alternatives | Why this one |
|---|---|---|
| Rewrite the Go backend in layers | keep the original code; switch to Node.js | the existing backend was Go; layers make it testable and clear; another language would have thrown work away |
| Standard `net/http` router | a web framework | Go's router supports methods and path parameters; no framework to learn, update or be locked into |
| One service, not microservices | split into services | the platform works well as one service; splitting adds network failures and many deployments, and would break the one-transaction checkout, with no real need |
| Frontend API layer with mappers | rewrite components to the backend's names | keeps the existing UI untouched; one place to adapt |
| API on the same address as the shop (`/api`) | a separate API domain | the host decides the store, no CORS, one frontend build for all stores |
| Lazy-load the admin panel | one bundle | shoppers do not download code they never use |

## Data and money

| Decision | Alternatives | Why this one |
|---|---|---|
| Money as decimals | floats | floats give rounding errors on prices |
| PostgreSQL row locks at checkout | check-then-update without locks | without locks two buyers can both get the last item |
| Orders keep copies of names, prices, address | look them up live | history must not change when the catalog changes |
| Soft deletes for products, bundles, accounts | real deletes | orders refer to them |
| Versioned SQL migrations | edit the database by hand; an ORM's auto-migrate | the structure changes with the code, reviewed and in order; nothing changes by surprise |

## Promotions and payments

| Decision | Alternatives | Why this one |
|---|---|---|
| One best price, no stacking (offers) | add discounts up | predictable prices; protects margins |
| Coupons on top of offers | never with offers; best of the two | simplest to explain; the admin controls the cost with the minimum order and the cap (decided by the project owner) |
| Coupons used once per customer by default | unlimited by default | the safe default; the admin can change it |
| Append-only debt ledger | an editable balance column | a full history of every charge and payment; mistakes are corrected visibly |
| Bundles as their own cart lines and tables | fake products | stock stays per real product; a bundle can change without touching products |

## Accounts and email

| Decision | Alternatives | Why this one |
|---|---|---|
| Email confirmation required for checkout only | block login too | customers can browse and fill a cart right away; only ordering needs a real email |
| Same answer for unknown emails (login, reset) | "no such account" | nobody can find out who is a customer |
| Token version for revocation | a token blacklist | one integer per user; works across every device at once |
| Email queue in the database | send during the request; Redis or a message queue | a slow mail provider must not break checkout; emails survive restarts; no extra service to run |
| Secret-link emails sent directly | queue them too | the queue would store the secret link |

## Many stores

| Decision | Alternatives | Why this one |
|---|---|---|
| One database, shared tables, RLS | a database or schema per store | no per-store infrastructure; a store is one row; RLS makes forgotten filters harmless |
| Store from the host name | a store id in the request | cannot be forged; works for subdomains and custom domains; browsers keep logins per host |
| The API runs as a non-owner role | run as the owner | owners skip RLS, so the policies would protect nothing |
| Store-aware composite foreign keys | plain foreign keys | foreign-key checks ignore RLS; this blocks cross-store links too |
| Separate platform admins table and tokens | a "super admin" role on users | a store login can never become a platform login |
| Invite email instead of a password | the platform admin types a password | nobody but the store admin ever knows their password |
| Page content as data with defaults | per-store component changes | one codebase; a new store looks complete; admins change only what they want |
| TXT record for domain ownership | trust the CNAME | pointing a domain proves nothing; only the owner can publish the token |
| On-demand certificates with an "ask" check | wildcard only; manual certificates | works for custom domains with no setup; the check blocks abuse |
| Per-store upload folders + ownership check | one shared folder | files manageable per store; no store can delete another's file |

## Operations

| Decision | Alternatives | Why this one |
|---|---|---|
| Caddy as the front door | Nginx + certbot | automatic HTTPS, including certificates on demand for custom domains, in a short config file |
| Distroless API image | a full Linux image | no shell or tools for an attacker; smaller |
| Production mode refuses unsafe settings | warnings only | a server that will not start is noticed in minutes; a weak secret might never be |
| Backups verified by reading back, plus a restore test | trust `pg_dump` | a backup that was never restored is a hope, not a backup |
