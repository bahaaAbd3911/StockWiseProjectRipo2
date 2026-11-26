# 1. Overview

## What StockWise is

StockWise is an **online-shop platform**. One codebase and one running
installation serve **many independent stores**. Each store has its own
address, products, customers, orders, prices, coupons, logo, colors and page
content, and no store can see or change another store's data.

A new store is **ready the moment it is created**. It needs no new database,
no new server and no new deployment. A platform admin fills in a form, the
store's admin receives an email, chooses a password and starts selling.

Everything is bilingual (English and Arabic, with right-to-left layout for
Arabic) and has a light and a dark theme.

## Who uses it

| Who | Where | What they do |
|---|---|---|
| **Customers** | the store's address, e.g. `ahmad.myplatform.com` or the store's own domain `ahmadstore.com` | browse, search, buy (cash on delivery or pay later), use coupons, buy bundles, track and cancel orders, review products, keep favorites, join waiting lists, see their balance |
| **Store admins** | the same address, under `/admin` | manage products, categories, orders, customers, coupons, offers, bundles, customer debts, the waiting list, the store's look and content, and its domains; see the dashboard |
| **Platform admins** | `admin.myplatform.com` | create stores (with their first admin), edit, suspend and reactivate them |

## The big picture

```
   ahmad.myplatform.com   omar.myplatform.com   ahmadstore.com   admin.myplatform.com
            \                    |                    |                /
             v                   v                    v               v
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Caddy (web server)                                                       │
   │  - HTTPS certificate for every host (only hosts the API confirms)        │
   │  - serves the React app: the same files for every store                  │
   │  - sends /api/* and /uploads/* to the Go API, keeping the host name      │
   └──────────────────────────────────┬───────────────────────────────────────┘
                                      v
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Go API                                                                   │
   │  1. which store?    from the host name                                   │
   │  2. who is it?      the login token, which must belong to that store     │
   │  3. what is allowed?  customer / store admin / platform admin            │
   │  4. business rules, then SQL                                             │
   │  + a background worker: sends queued emails and debt reminders           │
   └──────────────────────────────────┬───────────────────────────────────────┘
                                      v  role stockwise_app, app.store_id set
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ PostgreSQL: one database for all stores                                  │
   │  every row has store_id; row-level security lets a connection see        │
   │  only the rows of its current store                                      │
   └──────────────────────────────────────────────────────────────────────────┘
```

The most important idea: **the React app never knows a store id.** It calls
`/api/...` on the address it was opened from, and the backend works out the
store from that address. That is why one build of the frontend serves every
store and every custom domain, and why a customer cannot pretend to be in
another store by editing a request.

## The technology

| Part | Technology | Why |
|---|---|---|
| Backend | Go 1.27, standard `net/http` router, `pgx` PostgreSQL driver, `shopspring/decimal`, `golang-jwt`, `bcrypt` | the original backend was Go; the standard library covers routing, so there is no framework to learn or update |
| Database | PostgreSQL 18 | transactions, row locks, row-level security and constraints, which this project depends on |
| Frontend | React 19, Vite 8, react-router 7 | the teammate's template was React; Vite builds it fast |
| Web server | Caddy 2 | automatic HTTPS, including certificates on demand for custom domains |
| Packaging | Docker and Docker Compose | the whole platform runs on one server with one command |

## Where to go next

- How it was built over time: [02-history.md](02-history.md)
- How stores are kept apart: [05-multi-tenancy.md](05-multi-tenancy.md)
- Why each big choice was made: [24-decisions.md](24-decisions.md)
