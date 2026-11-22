# Going to production - checklist

The steps to put the API on a server, in order. Background for every step is
in [GUIDE.md](GUIDE.md) (sections 8-10).

## 1. Before you start

- [ ] A server (VPS) with Docker installed, ports **80** and **443** open.
  On Render / Railway / Fly.io, skip Caddy and follow "Option B" in GUIDE.md §9.
- [ ] Two domain names, e.g. `shop.example.com` (frontend) and
  `api.example.com` (API). Add a DNS **A** record (and **AAAA** for IPv6) for
  `api.example.com` pointing at the server.
- [ ] An SMTP provider (Brevo, Mailgun, Amazon SES, Postmark...). Verify your
  domain there and add the **SPF** and **DKIM** records it gives you, or
  emails will land in spam.
- [ ] The frontend has the pages listed in GUIDE.md §13 (`/verify-email`,
  `/reset-password`, `/orders/{id}`, `/account/debt`, `/products/{id}`).

## 2. Configure

```bash
git clone <repo> && cd <repo>/backend
cp .env.production.example .env.production
```

Fill in every `<...>`:

| Setting | How |
|---|---|
| `JWT_SECRET` | `openssl rand -base64 48` (PowerShell: `[Convert]::ToBase64String((1..48 \| % { [byte](Get-Random -Max 256) }))`) |
| `POSTGRES_PASSWORD` | `openssl rand -hex 24` (letters and digits only - it goes into a URL) |
| `API_DOMAIN`, `PUBLIC_URL` | `api.example.com`, `https://api.example.com` |
| `FRONTEND_URL`, `CORS_ALLOWED_ORIGINS` | `https://shop.example.com` (no trailing slash, no path) |
| `SMTP_*` | from your provider; `SMTP_FROM` on your verified domain |
| `CURRENCY`, `STORE_NAME` | as customers should see them |

Never commit `.env.production` (it is in `.gitignore`).

## 3. Start

```bash
docker compose --env-file .env.production up -d --build
docker compose ps                  # db healthy, migrate exited 0, api healthy, caddy running
docker compose logs api            # look for: msg=starting env=production
curl https://api.example.com/api/health
```

If the API refuses to start, `docker compose logs api` prints the full list
of settings to fix. Warnings (e.g. about the database) are logged but do not
stop it.

## 4. First admin

```bash
# register through the frontend first, then:
docker compose exec db psql -U shop_api -d ecommerce \
  -c "UPDATE users SET role = 'admin' WHERE email = 'you@example.com';"
```

## 5. Check

- [ ] `https://api.example.com/api/health` → `{"database":"ok","status":"ok"}`, padlock in the browser
- [ ] `http://api.example.com/api/health` redirects to HTTPS
- [ ] Register → the confirmation email arrives (check spam) → the link confirms
- [ ] Forgot password → the email arrives → reset works
- [ ] Place an order → "Order received" email; change its status → status email
- [ ] Upload a product image in the admin → it shows in the shop
- [ ] `SELECT COUNT(*) FROM email_outbox WHERE sent_at IS NULL;` stays near 0

## 6. Backups (do not skip)

```bash
# daily database dump, e.g. from cron at 03:00
docker compose exec -T db pg_dump -U shop_api -Fc ecommerce > backup-$(date +%F).dump
```

- Copy the dumps **off the server** (S3, another machine) and keep ~30 days.
- Also back up the `uploads` volume (`docker run --rm -v stockwise_uploads:/u -v $PWD:/b alpine tar czf /b/uploads.tgz -C /u .`).
- **Test a restore once**: `pg_restore -U shop_api -d ecommerce_restore_test backup.dump`.
- Managed databases (Neon, Supabase, RDS) have automatic backups: turn on point-in-time recovery.

## 7. Updating

```bash
git pull
docker compose --env-file .env.production up -d --build
```

Migrations run automatically before the new API starts. Take a backup first
when a release adds a migration.

## 8. Monitoring (recommended)

- An uptime check on `https://api.example.com/api/health` every minute
  (UptimeRobot, Better Stack, ...).
- Logs are JSON on stdout: `docker compose logs -f api`, or ship them to a log
  service. Search for `level=ERROR`.
- Watch stuck emails: `SELECT kind, to_address, attempts, last_error FROM email_outbox WHERE sent_at IS NULL;`
