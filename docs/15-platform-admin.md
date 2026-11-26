# 15. The platform super admin

The people who run the platform itself, not any one store.

## 15.1 Where and who

- **Where:** only on `PLATFORM_ADMIN_HOST` (default `admin.<PLATFORM_DOMAIN>`,
  `off` disables it).
  - The `/api/platform/*` API answers 404 on every other host.
  - That host serves nothing but the platform API: store endpoints are
    refused there.
  - The React app sees the host and shows the platform console instead of a
    shop (`main.jsx`).
- **Who:** rows of `platform_admins`. They are separate from store users,
  with their own login tokens from a different issuer.
  - A store admin's token is refused by the platform API.
  - A platform token is refused by every store.

**There is no sign-up for platform admins.** The first one is created on the
server:

```
go run ./cmd/platform create-admin -email you@example.com -first Sara -last Ahmad
# with Docker:
docker compose --env-file .env.production exec api /app/platform create-admin -email you@example.com
```

It prints a random password once.

**Why separate from store users?** If "super admin" were a role on the
`users` table, a bug in a store's user management could promote a store
account to run the whole platform. With separate tables and token issuers,
a store login can never become a platform login.

## 15.2 Creating a store

`POST /api/platform/stores {slug, name, currency?, timezone?, admin: {first_name, last_name, email}}`
creates, **in one transaction**:

1. the store record;
2. its settings row (with its name);
3. its admin account: role admin, email confirmed, **no usable password**;
4. an invite token.

Then the admin gets an email, sent as the store, with a link to **the
store's own** reset-password page (valid 72 hours) to choose a password.

The store is live at `<slug>.<PLATFORM_DOMAIN>` at once, with no new
database and no deployment.

- **Why an invite instead of a password?** Nobody but the store's admin
  ever knows their password, not even the platform admin.
- **Slugs** must be valid DNS labels. `admin`, `www`, `api`, `mail` and
  other names are reserved, so a store cannot take the platform's own
  addresses.

## 15.3 Managing stores

| Action | Effect |
|---|---|
| list | each store with its numbers: products, customers, orders, revenue; search and status filter |
| edit | name, currency, time zone (the slug never changes, because addresses and links depend on it) |
| suspend | everyone gets 403 "this store is currently unavailable"; **nothing is deleted**; no new certificates |
| reactivate | back to normal |
| re-invite | a new invite link; the old one stops working |

Platform requests run in **platform mode**, so they can see every store's
rows for the counts. Store requests never can ([05](05-multi-tenancy.md)).
