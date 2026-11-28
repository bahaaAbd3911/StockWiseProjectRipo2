# 20. Running it on your computer

## Requirements

- Go (the version in `backend/go.mod`)
- Node.js and npm
- PostgreSQL, running locally (or a cloud database such as Neon)

## The development database (PostgreSQL in Docker)

The recommended local database runs in Docker (`backend/docker-compose.dev.yml`):
nothing to install on Windows, the same PostgreSQL 18 as production and CI,
and it can be reset without touching anything else.

```powershell
cd backend
# once: add a random password to .env
#   DEV_DB_PASSWORD=<random letters and digits>
#   DATABASE_URL=postgres://postgres:<that password>@localhost:5433/ecommerce?sslmode=disable
docker compose -f docker-compose.dev.yml up -d    # start (restarts itself after a reboot)
go run ./cmd/migrate up                           # create / update the tables
```

- It listens only on this computer, on port **5433**. Port 5432 is left
  free for a PostgreSQL installed on Windows.
- The data lives in the Docker volume `stockwise-dev_pgdata`. `stop` and
  `up` keep it; deleting the volume deletes it.
- Backup:

  ```
  docker exec stockwise-dev-db-1 pg_dump -U postgres --format=custom ecommerce > dev.dump
  ```

- The data from the earlier Windows PostgreSQL (port 5432) was copied into
  it on 2026-11-28. Every table's row count matched, and so did the
  isolation policies and permissions. The old database was left untouched,
  and its setting is kept as a comment in `.env`.

**If the site is slow on Windows:** measure where the time goes before
blaming the database. On this project's PC (2026-11-28):

- inside Docker's network, queries took under 1 ms;
- from Windows programs to `localhost`, 42% of them stalled for up to
  3 seconds, whichever database was used.

The cause is on the Windows side. The usual suspects are a loopback
capture driver (the **Npcap Loopback Adapter**, installed with
Wireshark/Nmap) and antivirus scanning of database files and local
traffic. Disable or uninstall them (administrator rights needed) and
measure again.

## One shop

```powershell
cd backend
copy .env.example .env        # set DATABASE_URL, and a random JWT_SECRET of 32+ characters
go run ./cmd/migrate up       # create / update the database
go run ./cmd/api              # the API on http://localhost:8080

# in a second terminal
cd Front-End
npm install
npm run dev                   # the shop on http://localhost:5173
```

From the project root, `npm run api` and `npm run dev` do the same.

- `localhost` is the `main` store.
- **Emails are printed in the API's terminal** (`MAIL_DRIVER=log`). Copy
  confirmation and reset links from there.
- **The first store admin:** register on the shop, then in PostgreSQL run
  `UPDATE users SET role = 'admin' WHERE store_id = 1 AND email = 'you@example.com';`,
  then log in again.

## Several stores and the platform console

1. Add `PLATFORM_DOMAIN=localhost` to `backend/.env` and restart the API.
2. Create a platform admin (it prints a password):

   ```powershell
   go run ./cmd/platform create-admin -email you@example.com
   ```

3. Open `http://admin.localhost:5173`, sign in, and create a store `ahmad`.
4. The invite link is printed in the API terminal. Open it to set the store
   admin's password.
5. Open `http://ahmad.localhost:5173`.

Browsers send every `*.localhost` name to your own computer, and the Vite
proxy keeps the name, so the backend sees which store was opened.

## When something does not work

| Symptom | Likely cause |
|---|---|
| API exits at start with a list of settings | fix what it lists; production mode is strict on purpose |
| "relation ... does not exist" | run `go run ./cmd/migrate up` |
| every page says "store not found" | `PLATFORM_DOMAIN` is set but you opened `localhost`; open `<slug>.localhost:5173`, or set `DEFAULT_STORE=main` |
| checkout says "please confirm your email" | open the confirmation link printed in the API terminal |
| 429 while testing | the rate limits; wait for `Retry-After`, or set `RATE_LIMIT_ENABLED=false` locally (never in production) |
