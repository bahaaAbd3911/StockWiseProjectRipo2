# 3. Repository layout

```
Project/
├── .github/workflows/ci.yml      GitHub Actions: tests, lint, build, scan on every push and weekly
├── .github/dependabot.yml        weekly pull requests for dependency updates
├── docs/                         this documentation
├── package.json                  shortcuts run from the root: npm run dev / api / build / lint
├── Front-End/                    the React app (Vite)
│   ├── index.html                the single HTML page; loads fonts and icons
│   ├── vite.config.js            dev server; passes /api and /uploads to the backend
│   └── src/
│       ├── main.jsx              entry: platform console on admin.* hosts, the shop everywhere else
│       ├── App.jsx               the shop's routes (client, admin, account pages, email links)
│       ├── api/                  every call to the backend, and the field conversions
│       ├── context/              app-wide state: session, store settings, catalog, cart
│       ├── hooks/                small reusable pieces of logic
│       ├── layouts/              the shop layout and the admin layout
│       ├── Pages/                screens: client/, admin/, account/, platform/
│       ├── components/           building blocks: client/, admin/ (UI kit in admin/ui), common/
│       ├── data/, constants/     default texts, settings defaults, labels
│       └── utils/                formatting and validation helpers
└── backend/                      the Go API
    ├── cmd/
    │   ├── api/                  the server program
    │   ├── migrate/              applies database migrations (status / up)
    │   └── platform/             creates a platform admin (create-admin)
    ├── internal/                 all application code (Go forbids importing it from outside)
    ├── migrations/               numbered SQL changes 000001 ... 000009, each with up and down
    ├── tests/                    end-to-end tests against a throwaway database
    ├── deploy/
    │   ├── Caddyfile             the web server: HTTPS, routing, security headers
    │   ├── Dockerfile.web        builds the React app and puts it in a Caddy image
    │   ├── backup.sh             daily database + uploads backup
    │   └── restore.sh            test a backup, or restore it
    ├── Dockerfile                the API image
    ├── docker-compose.yml        the whole platform on one server: db, migrate, api, web
    ├── .env.example              development settings, every one commented
    └── .env.production.example   server settings
```

## Why the split looks like this

- **`backend/internal/`**: Go refuses imports of `internal` packages from
  outside the module, so the application code cannot be used by accident as
  a library.
- **`cmd/` has three small programs** instead of one with many flags. Each
  does one job, and the Docker image contains all three.
- **`migrations/` lives next to the code**: a code change that needs a new
  column ships with the SQL that adds it, in the same commit.
- **`deploy/` holds everything server-specific**, so the application code
  does not know whether it runs in Docker, behind Caddy or alone.
- **The frontend's `api/` folder is the only place that talks to the
  backend.** Components never call `fetch` themselves, so a change in the
  backend is fixed in one folder.

The inside of `backend/internal/` is described package by package in
[06-backend-architecture.md](06-backend-architecture.md), the inside of
`Front-End/src/` in [16-frontend.md](16-frontend.md).
