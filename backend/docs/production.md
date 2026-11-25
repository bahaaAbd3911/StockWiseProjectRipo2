# Going to production - checklist

The steps to put the whole platform (every store, the platform admin, custom
domains) on one server, in order. Background: [../../DOCUMENTATION.md](../../DOCUMENTATION.md)
sections 6 and 13, and [GUIDE.md](GUIDE.md) sections 8-10.

## 1. Before you start

- [ ] A server (VPS) with Docker installed, ports **80** and **443** open.
- [ ] The platform's domain, e.g. `myplatform.com`. At its DNS provider add
  **A** records (and **AAAA** for IPv6) pointing at the server for:
  - `myplatform.com` (the default store / landing page)
  - `*.myplatform.com` (every store, and `admin.myplatform.com`)
- [ ] An SMTP provider (Brevo, Mailgun, Amazon SES, Postmark...). Verify your
  domain there and add the **SPF** and **DKIM** records it gives you, or
  emails will land in spam.

## 2. Configure

```bash
git clone <repo> && cd <repo>/backend
cp .env.production.example .env.production
```

Fill in every `<...>`:

| Setting | How |
|---|---|
| `PLATFORM_DOMAIN` | `myplatform.com` |
| `FRONTEND_URL`, `PUBLIC_URL`, `CORS_ALLOWED_ORIGINS` | `https://myplatform.com` |
| `JWT_SECRET` | `openssl rand -base64 48` (PowerShell: `[Convert]::ToBase64String((1..48 \| % { [byte](Get-Random -Max 256) }))`) |
| `POSTGRES_PASSWORD` | `openssl rand -hex 24` (letters and digits only - it goes into a URL) |
| `SMTP_*` | from your provider; `SMTP_FROM` on your verified domain (each store's emails show the store's name with this address) |
| `ACME_EMAIL` | optional: Let's Encrypt certificate notices |

Never commit `.env.production` (it is in `.gitignore`).

## 3. Start

```bash
docker compose --env-file .env.production up -d --build
docker compose ps                  # db healthy, migrate exited 0, api healthy, web running
docker compose logs api            # look for: msg=starting env=production
curl https://myplatform.com/api/health
```

If the API refuses to start, `docker compose logs api` prints the full list
of settings to fix. Warnings are logged but do not stop it (the warning about
`DEFAULT_STORE` with `PLATFORM_DOMAIN` is expected when the bare domain
shows the `main` store).

## 4. First platform admin, first store

```bash
docker compose --env-file .env.production exec api /app/platform create-admin -email you@example.com -first You -last Name
```

It prints a password once. Open `https://admin.myplatform.com`, sign in,
create a store (address, name, its admin's name and email). The store's admin
gets an invitation email and the store is live at `https://<address>.myplatform.com`.

The existing `main` store keeps its data. To give it an admin, register on
`https://myplatform.com` and run:
`docker compose exec db psql -U shop_api -d ecommerce -c "UPDATE users SET role = 'admin' WHERE store_id = 1 AND email = 'you@example.com';"`

## 5. Check

- [ ] `https://myplatform.com/api/health` → `{"database":"ok","status":"ok"}`, padlock in the browser
- [ ] `http://myplatform.com` redirects to HTTPS
- [ ] `https://admin.myplatform.com` shows the platform login
- [ ] A new store opens at its address; `https://nosuchstore.myplatform.com/api/products` → 404
- [ ] Register on a store → the confirmation email (sent as the store) links to that store
- [ ] Place an order → "Order received" email; change its status → status email
- [ ] Upload a product image → it shows in the shop
- [ ] A custom domain: add it in the store's admin (Domains), create the TXT and CNAME records, verify, open it (the certificate is issued on the first visit)
- [ ] `GET /api/admin/emails/summary` shows `failed: 0` and a small `pending`

### Many stores: one wildcard certificate

Certificates are issued on demand per host. Let's Encrypt allows **50 per
week per registered domain**, so with many store subdomains switch them to one
wildcard certificate for `*.myplatform.com`. That needs Caddy built with your
DNS provider's module (e.g. `xcaddy build --with github.com/caddy-dns/cloudflare`)
and an API token; add a site block `*.myplatform.com { tls { dns cloudflare {env.CF_API_TOKEN} } ... }`
with the same handlers as the `https://` block. Custom domains keep using
on-demand certificates.

## 6. Backups (do not skip)

1. Somewhere **off the server** for the copies. For example a Cloudflare R2 or
   S3 bucket: install rclone (`curl https://rclone.org/install.sh | sudo bash`)
   and run `rclone config` to add it as a remote, e.g. `r2`.
2. Run the first backup by hand and check it:
   ```bash
   BACKUP_REMOTE=r2:stockwise-backups deploy/backup.sh
   ls -lh /var/backups/stockwise/
   ```
3. **Prove the backup restores** (does not touch live data):
   ```bash
   deploy/restore.sh --test /var/backups/stockwise/stockwise-db-<date>.dump
   ```
4. Schedule it daily (`crontab -e`):
   ```
   0 3 * * * cd /opt/stockwise/backend && BACKUP_REMOTE=r2:stockwise-backups deploy/backup.sh >> /var/log/stockwise-backup.log 2>&1
   ```
5. Repeat step 3 every month or so, and read the log now and then
   (`grep FAILED /var/log/stockwise-backup.log`).

Restoring for real (replaces the live data, asks for confirmation):
`deploy/restore.sh --live <db.dump> [<uploads.tar.gz>]`.

Managed databases (Neon, Supabase, RDS) also have their own automatic
backups: turn on point-in-time recovery there as well.

## 7. Updating

```bash
git pull
docker compose --env-file .env.production up -d --build
```

Migrations run automatically before the new API starts. Take a backup first
when a release adds a migration.

## 8. Monitoring (recommended)

- An uptime check on `https://myplatform.com/api/health` every minute
  (UptimeRobot, Better Stack, ...).
- Logs are JSON on stdout: `docker compose logs -f api`, or ship them to a log
  service. Search for `level=ERROR`.
- Watch the email queue: `GET /api/admin/emails/summary` (as an admin). A
  growing `failed` count or an old `oldest_pending_at` means emails are not
  going out - check the SMTP settings, then `POST /api/admin/emails/retry-failed`.
