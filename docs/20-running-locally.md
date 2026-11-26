# 20. Running it on your computer

## Requirements

- Go (the version in `backend/go.mod`)
- Node.js and npm
- PostgreSQL, running locally (or a cloud database such as Neon)

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
