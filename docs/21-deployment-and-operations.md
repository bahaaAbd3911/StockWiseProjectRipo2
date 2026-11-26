# 21. Deployment and operations

## 21.1 What runs on the server

`backend/docker-compose.yml` runs the whole platform on one server:

```
db (PostgreSQL 18, volume db-data) ──healthy──► migrate (runs "migrate up", exits)
                                                     │ completed
                                                     ▼
                       web (Caddy + the built React app) :80 / :443  ──►  api :8080 (volume uploads)
```

**Only `web` is reachable from outside.** The database and the API are on
the internal Docker network.

| Image | Built from | Contents |
|---|---|---|
| API | `backend/Dockerfile` | stage 1 compiles static `api`, `migrate`, `platform` binaries; stage 2 is `distroless/static:nonroot`: no shell, no package manager, an unprivileged user, only the binaries, the SQL migrations and CA certificates; a `HEALTHCHECK` runs `api healthcheck` |
| web | `backend/deploy/Dockerfile.web` | stage 1 builds the React app with Node 22; stage 2 is `caddy:2-alpine` with the files in `/srv` and the Caddyfile |

**Why distroless?** An attacker who found a hole in the API would land in a
container with no shell and no tools to work with.

## 21.2 Caddy (`backend/deploy/Caddyfile`)

- **`http://`** → 308 redirect to HTTPS (308 keeps POST as POST).
- **`https://`** for every host, with on-demand certificates (only after
  `/internal/tls-check` says yes):
  - `/internal/*` → 404 from outside;
  - `/api/*` → the API, keeping the host name (the API needs it to find the
    store);
  - `/uploads/*` → the API;
  - everything else → the React app. Unknown paths load `index.html`, so
    app routes like `/client/home` work on reload.
- **Security headers on the pages:** CSP, `X-Frame-Options: DENY`,
  `nosniff`, referrer policy, HSTS ([07](07-accounts-and-security.md)).
- **Caching:**
  - `/assets/*` has content-hashed names, so it is cached for a year
    (`immutable`);
  - every other page is `no-cache`, so a new release shows up at once.

**Note on the CSP:** it allows `connect-src 'self'`, which fits the normal
setup where the API is under `/api` on the same address. If you ever point
`VITE_API_URL` at another domain, add that domain to `connect-src`.

## 21.3 Going live, step by step

1. **Server:** a VPS with Docker, ports 80 and 443 open.
2. **DNS:** `A` (and `AAAA`) records pointing at the server for
   `myplatform.com` and `*.myplatform.com` (every store and
   `admin.myplatform.com`).
3. **Email:** an SMTP provider (Brevo, Mailgun, Amazon SES, Postmark...).
   Verify your domain there and add its **SPF** and **DKIM** records, or
   emails will land in spam.
4. **Configure:**

   ```bash
   cd backend
   cp .env.production.example .env.production
   ```

   Fill in every `<...>`:
   - `PLATFORM_DOMAIN`;
   - `FRONTEND_URL`, `PUBLIC_URL`, `CORS_ALLOWED_ORIGINS` (https);
   - `JWT_SECRET`, `POSTGRES_PASSWORD` ([18](18-configuration.md));
   - `SMTP_*`;
   - `ACME_EMAIL`.
5. **Start:**

   ```bash
   docker compose --env-file .env.production up -d --build
   docker compose ps          # db healthy, migrate exited 0, api healthy, web running
   docker compose logs api    # "starting env=production"
   ```

   If the API refuses to start, its log lists every setting to fix.
6. **First platform admin** (prints a password once):

   ```bash
   docker compose --env-file .env.production exec api /app/platform create-admin -email you@example.com
   ```

7. Open `https://admin.myplatform.com`, sign in, create the first store.

## 21.4 Checks after going live

- [ ] `https://myplatform.com/api/health` → `{"database":"ok","status":"ok"}`, padlock shown
- [ ] `http://myplatform.com` redirects to HTTPS
- [ ] `https://admin.myplatform.com` shows the platform login
- [ ] a new store opens at its address; `https://nosuchstore.myplatform.com/api/products` → 404
- [ ] register on a store → the confirmation email links to that store
- [ ] place an order → "Order received" email; change its status → status email
- [ ] upload a product image → it shows in the shop
- [ ] open the shop with the browser's developer console: no "Content Security Policy" errors
- [ ] a custom domain: add, create the records, verify, open it
- [ ] `GET /api/admin/emails/summary` shows `failed: 0`

## 21.5 Many stores: one wildcard certificate

Let's Encrypt issues at most **50 certificates per week per registered
domain**. With many store subdomains, switch them to one wildcard
certificate for `*.myplatform.com`:

1. Build Caddy with your DNS provider's module, e.g.
   `xcaddy build --with github.com/caddy-dns/cloudflare`.
2. Add a site block
   `*.myplatform.com { tls { dns cloudflare {env.CF_API_TOKEN} } ... }`
   with the same handlers as the `https://` block.

Custom domains keep using on-demand certificates.

## 21.6 Backups (do not skip)

`backend/deploy/backup.sh`, run daily from cron:

1. **Dump the database** with `pg_dump` in custom format (compressed, and
   restorable table by table). It runs inside the database container, so
   nothing needs to be installed on the server; with a managed database it
   runs in a temporary postgres container.
2. **Read the dump back** with `pg_restore --list` before keeping it, so a
   broken dump is never mistaken for a backup.
3. **Archive the uploads volume.** It is written to a `.partial` file first
   and renamed only when complete.
4. **Delete** files older than `KEEP_DAYS` (30).
5. With `BACKUP_REMOTE` set (an rclone remote: Cloudflare R2, S3, Backblaze,
   Google Drive), **copy everything off the server**. A backup on the same
   disk does not survive the disk.

Files are readable by their owner only (`umask 077`), because they contain
customer data.

```bash
BACKUP_REMOTE=r2:stockwise-backups deploy/backup.sh       # first run, by hand
deploy/restore.sh --test /var/backups/stockwise/stockwise-db-<date>.dump
# cron: 0 3 * * * cd /opt/stockwise/backend && BACKUP_REMOTE=r2:stockwise-backups deploy/backup.sh >> /var/log/stockwise-backup.log 2>&1
```

- `restore.sh --test` restores into a separate database, prints row counts
  of the main tables, and drops it. It proves the backup works without
  touching live data. Repeat it monthly.
- `restore.sh --live <dump> [uploads.tar.gz]` replaces the live data:
  1. it asks you to type `restore`;
  2. it stops `api` and `web`;
  3. it restores;
  4. it starts everything again (pending migrations run first).

Both scripts were run for real during development: a backup, a test
restore, and a live restore.

## 21.7 Updating

```bash
git pull
docker compose --env-file .env.production up -d --build
```

Migrations run before the new API starts. Take a backup first when a
release adds a migration.

## 21.8 Monitoring

- An uptime check on `/api/health` every minute (UptimeRobot, Better
  Stack...).
- Logs are JSON on stdout: `docker compose logs -f api`. Search for
  `level=ERROR`.
- Watch the email queue (`/api/admin/emails/summary`): a growing `failed`
  count or an old `oldest_pending_at` means emails are not going out. Check
  the SMTP settings, then `POST /api/admin/emails/retry-failed`.
- `grep FAILED /var/log/stockwise-backup.log` now and then.

## 21.9 Without Docker

Build with `go build ./cmd/api ./cmd/migrate ./cmd/platform`. Then:

1. run `migrate up`;
2. run `api` under systemd behind Caddy or Nginx;
3. serve `Front-End/dist` with the same rules as the Caddyfile.

Alternatively, with `TLS_MODE=auto` the API can get its own Let's Encrypt
certificates when it runs alone on a server.
