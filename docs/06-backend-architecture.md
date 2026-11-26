# 6. Backend architecture

## 6.1 Layers

```
middleware  →  routes  →  handlers  →  services  →  repository  →  PostgreSQL
(every         (who may     (HTTP in/out)  (business     (SQL,          (constraints:
 request)       call what)                  rules)        transactions)   last safety net)
```

| Layer | May use | Must never contain |
|---|---|---|
| handlers | HTTP, JSON, services | SQL, business rules |
| services | models, validators, repositories, notify, storage | HTTP status codes, SQL |
| repository | SQL, models | HTTP |

**Why layers?**

- Each file has one job. A change in a shop rule touches a service; a change
  in a query touches a repository; nothing else moves.
- Each part can be tested on its own.
- A new developer knows where to look: "what is the rule?" is in
  `services/`, "what is the SQL?" is in `repository/`.

## 6.2 The packages in `backend/internal`

| Package | Job |
|---|---|
| `tenant` | the request's store; the resolver that turns a host name into a store (cached 30 s) |
| `database` | connection pool, the store-aware connection wrapper (`ConnectApp`), `WithTx` transactions, `Querier` (pool or transaction) |
| `config` | the only package that reads environment variables; defaults, checks, production refusals |
| `apperr` | the error vocabulary: kinds that map to status codes; database error translation |
| `httpx` | JSON in and out (size limits, clear messages), `WriteError`, path ids, pagination, client IP |
| `auth` | login tokens (store tokens and platform tokens, different issuers), the caller's `Identity` |
| `middleware` | store detection, authentication, admin / verified / platform checks, CORS, rate limits, security headers, logging, crash recovery |
| `routes` | builds everything and lists every endpoint with its access level; reading it shows the whole API |
| `handlers` | one file per area: read input → call one service → write JSON |
| `services` | the rules per area; queue emails after successful actions |
| `repository` | the SQL per area |
| `models` | data shapes and pure rules (which order status may follow which) |
| `validators` | clean input (trim, lower-case emails, empty → NULL) and check it, reporting every problem at once |
| `pricing` | the only place prices and discounts are calculated |
| `notify` | email texts (English + Arabic, per store), the email queue and its worker, per-store branding |
| `mail` | sending: `LogMailer` (prints, development) or `SMTPMailer` |
| `storage` | uploaded files: local disk or S3, one folder per store |
| `images` | checks that an upload really is an image; strips GPS and other metadata |
| `ratelimit` | in-memory token buckets |
| `server` | HTTP / HTTPS listeners, HTTP → HTTPS redirect, graceful stop |
| `migrate` | applies the SQL migrations in order, records them in `schema_migrations` |

## 6.3 The programs in `backend/cmd`

| Program | What it does |
|---|---|
| `api` | loads settings (production refuses to start on unsafe ones), connects, builds the routes, serves. On Ctrl+C / SIGTERM it stops accepting requests, gives running ones 10 s, waits for account emails being sent and stops the worker. `api healthcheck` is used by Docker. |
| `migrate` | `status` (read-only) and `up` |
| `platform` | `create-admin -email ... -first ... -last ...` creates a platform admin and prints a random password once |

## 6.4 Rules every file follows

**Exact money.** `decimal.Decimal` in Go, `numeric(10,2)` in the database,
sent as JSON numbers with at most 2 decimals. Never floats.

**All-or-nothing writes.** Anything touching more than one row runs inside
`database.WithTx`. If any step fails, nothing is saved.

**Row locks against races.** `SELECT ... FOR UPDATE` makes concurrent
requests wait their turn where it matters:

- the last unit of stock;
- a customer's credit;
- the last use of a coupon;
- the one default address and the one primary image;
- single-use tokens;
- the email queue (`SKIP LOCKED`).

**Not yours = not found.** Every query on personal data filters by user
(and RLS filters by store). Someone else's order id answers 404, as if it
did not exist, and does not even reveal that it exists.

**Soft deletes.** Products, bundles and accounts are deactivated, not
deleted, because orders refer to them. Orders keep **copies** of names,
prices, colors and the address, so history never changes when the catalog
does.

**One rule, one place.** Each rule lives in one function used by every
screen, so two screens can never disagree:

| Rule | Where |
|---|---|
| the price | `pricing.BestPrice` |
| can this line be bought | `cartLine.problem` and `bundleProblem` |
| is this offer live | `offerAppliesSQL` |
| which order status may follow which | `models.CanTransitionOrder` |

**Secrets are hashed.** Passwords with bcrypt. Emailed links (reset,
confirmation, store invite) contain random 32-byte tokens; only a SHA-256 of
the token is stored. Someone who reads the database cannot log in or use a
link.

**SQL is always parameterised.** Values go through `$1, $2 ...`
placeholders. The only text joined into SQL is fixed fragments written in
the code (column lists, a whitelisted sort order), never user input.

**Typed errors, not strings**: see [04-request-lifecycle.md](04-request-lifecycle.md).

## 6.5 Validation

`validators.Checker` collects every field problem before answering:

```json
{"error": "validation failed", "fields": {"email": "is required", "price": "must be more than 0"}}
```

Each `Validate...` function does two steps:

1. **Normalise**: trim spaces, lower-case emails, upper-case coupon codes,
   empty strings → NULL.
2. **Check** against the database's limits: lengths in characters, money
   ≤ 99,999,999.99 with 2 decimals, phone, email, URL and hex-color formats.

Checking before the database gives a clear 400 naming the field instead of
a database error, and the frontend forms show the message under the right
input.
