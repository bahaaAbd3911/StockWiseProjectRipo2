# 18. Configuration

All settings are environment variables. Locally they come from
`backend/.env`; on a server from `.env.production`. The commented templates
are:

- `backend/.env.example` (development);
- `backend/.env.production.example` (a server).

**Never commit a filled-in `.env` or `.env.production`.** They contain the
database password and the login-signing secret.

## 18.1 Every setting

| Group | Setting | Meaning |
|---|---|---|
| Basics | `APP_ENV` | `development` (default) or `production` (18.2) |
| | `APP_PORT` | port the API listens on (8080) |
| | `LOG_FORMAT` | `text` or `json` (json is the production default) |
| Database | `DATABASE_URL` | PostgreSQL connection string |
| | `DB_MAX_CONNS` | maximum open connections (25) |
| Logins | `JWT_SECRET` | secret that signs login tokens; **at least 32 characters**, random |
| | `JWT_TTL` | how long a login lasts (24h) |
| Stores | `PLATFORM_DOMAIN` | stores live at `<slug>.PLATFORM_DOMAIN`; empty = one shop |
| | `DEFAULT_STORE` | store for any other host (`main`); empty = 404 |
| | `PLATFORM_ADMIN_HOST` | platform console host; empty = `admin.<PLATFORM_DOMAIN>`, `off` = none |
| | `STORE_URL` | a store's address in emails, with `{slug}` |
| Browser access | `CORS_ALLOWED_ORIGINS` | origins allowed to call the API |
| | `TRUST_PROXY` | read the client IP from `X-Forwarded-For`; **only behind a proxy** |
| | `RATE_LIMIT_ENABLED` | keep `true` (the numbers are in `config.DefaultRateLimits`) |
| HTTPS (API itself) | `TLS_MODE` | `off` (a proxy does HTTPS), `auto` (Let's Encrypt), `files` (own certificate) |
| | `TLS_DOMAINS`, `TLS_EMAIL`, `TLS_CACHE_DIR`, `TLS_CERT_FILE`, `TLS_KEY_FILE`, `TLS_PORT`, `TLS_HTTP_PORT` | for `auto` / `files` |
| The shop | `STORE_NAME` | fallback name (a store's own name comes from its record) |
| | `FRONTEND_URL` | the shop's address, for emailed links |
| | `CURRENCY` | fallback currency for emails |
| | `STORE_TIMEZONE` | when a dashboard day starts (`Asia/Amman`) |
| | `LOW_STOCK_THRESHOLD` | "low stock" level (5) |
| Email | `MAIL_DRIVER` | `log` (print) or `smtp` |
| | `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM` | the SMTP provider |
| Links & tokens | `PASSWORD_RESET_URL`, `PASSWORD_RESET_TTL` | reset page (default `FRONTEND_URL/reset-password`), 30m |
| | `EMAIL_VERIFY_URL`, `EMAIL_VERIFY_TTL` | confirmation page (default `FRONTEND_URL/verify-email`), 48h |
| | `REQUIRE_EMAIL_VERIFICATION` | checkout needs a confirmed email (`true`) |
| Queue | `EMAIL_POLL_INTERVAL` | how often the queue is checked (5s) |
| | `DEBT_REMINDER_INTERVAL` | reminder period (168h; 0 = off) |
| Uploads | `PUBLIC_URL` | the API's public address (images are served from it) |
| | `STORAGE_DRIVER` | `local` or `s3` |
| | `UPLOAD_DIR`, `MAX_UPLOAD_MB` | folder (`data/uploads`), size cap (5) |
| | `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_PUBLIC_URL`, `S3_INSECURE` | any S3-compatible bucket |
| Docker only | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | the database container |
| | `ACME_EMAIL` | **required**: Let's Encrypt's contact for certificate warnings; Caddy refuses to start without it |

## 18.2 Production mode

With `APP_ENV=production` the API **refuses to start** and prints every
problem at once when:

| Problem | Why it matters |
|---|---|
| `JWT_SECRET` is the example value or too predictable | anyone who knows it can sign in as any user, including admins |
| `CORS_ALLOWED_ORIGINS` is `*`, not https, or localhost | other websites could use visitors' logins |
| `FRONTEND_URL`, `PASSWORD_RESET_URL`, `EMAIL_VERIFY_URL` not public https | emailed links carry secrets |
| `PUBLIC_URL` / `S3_PUBLIC_URL` not public https | browsers would refuse or warn about images |
| `MAIL_DRIVER` is not `smtp` | the log driver writes reset links into the logs, and nobody receives mail |
| `RATE_LIMIT_ENABLED=false` | password guessing becomes unlimited |
| `S3_INSECURE=true` | uploads would travel unencrypted |

It **warns** but keeps running about:

- `sslmode=disable` to a remote database;
- the `postgres` superuser in `DATABASE_URL`;
- `JWT_TTL` over 7 days;
- links whose domain is not in the CORS list;
- `TLS_MODE=off` with `TRUST_PROXY=false` (probably a missing proxy
  setting);
- `TLS_MODE` on together with `TRUST_PROXY=true`.

**Why refuse instead of warn?** Each of these problems is invisible in
normal use and catastrophic when exploited. A server that will not start is
noticed in minutes; a weak secret might never be noticed.

## 18.3 Making secrets

```
openssl rand -base64 48          # JWT_SECRET
openssl rand -hex 24             # POSTGRES_PASSWORD (letters and digits: it goes into a URL)
```

PowerShell: `[Convert]::ToBase64String((1..48 | % { [byte](Get-Random -Max 256) }))`
