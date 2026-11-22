# StockWise backend (Go + PostgreSQL)

REST API for the StockWise shop: accounts with email confirmation, a
bilingual (English/Arabic) catalog with image uploads, cart, favorites,
addresses, checkout (cash on delivery or pay later), order tracking, reviews,
a waiting list, timed offers, bundles, customer debts, and email
notifications - built for production with HTTPS, rate limits and Docker.

**Start here: [docs/GUIDE.md](docs/GUIDE.md)** - the complete guide: how the
code is organised, the logic of every feature, configuration, HTTPS,
deployment, tests and the full API reference.

| Document | For |
|---|---|
| [docs/GUIDE.md](docs/GUIDE.md) | everything, from the architecture to the API reference |
| [docs/production.md](docs/production.md) | step-by-step server checklist (configure, start, check, back up) |
| [docs/offers-bundles-debts.md](docs/offers-bundles-debts.md) | offers, bundles and pay-later in depth |
| [docs/auth-security.md](docs/auth-security.md) | rate limits, login revocation, password reset in depth |

## Run it locally

Requirements: Go 1.27+, PostgreSQL 18.

```powershell
copy .env.example .env          # set DATABASE_URL and a JWT_SECRET of 32+ characters
go run ./cmd/migrate up         # create / update the schema (run after every pull)
go run ./cmd/api                # http://localhost:8080/api/health
```

In development `MAIL_DRIVER=log` prints every email (confirmation links,
reset links, order emails) in the terminal instead of sending it, and
uploaded images are stored in `data/uploads`.

First admin: register, then `UPDATE users SET role = 'admin' WHERE email = 'you@example.com';`

## Run it in production

```bash
cp .env.production.example .env.production      # fill in every <...>
docker compose --env-file .env.production up -d --build
```

PostgreSQL, migrations, the API and Caddy (automatic HTTPS) start in order.
With `APP_ENV=production` the API refuses unsafe settings and lists what to
fix. Details: [docs/production.md](docs/production.md).

## Tests

```powershell
go test ./...                                                   # unit tests
$env:INTEGRATION_DATABASE_URL = "postgres://USER:PASS@localhost:5432/postgres?sslmode=disable"
go test ./tests -v                                              # end-to-end, on a throwaway database
```
