# 16. The frontend (React)

## 16.1 Structure of `Front-End/src`

| Folder | Role |
|---|---|
| `main.jsx` | entry point: on the platform admin host it shows the platform console, anywhere else the shop |
| `App.jsx` | the shop's routes: `/client/*`, `/admin/*`, sign in/up, account pages, email-link forwards |
| `api/` | **the only place that talks to the backend** (16.2) |
| `context/` | app-wide state (16.3) |
| `hooks/` | reusable logic: `useLang` (language + `tr(en, ar)`), `useToast`, `useApiAction` (busy + toasts), `useShop`, `useCatalog`, `useStoreContent`, `useStoreTheme`, `usePagination`, `useFilters` ... |
| `layouts/` | `ClientLayout` (shop routes, the "confirm your email" banner), `AdminLayout` (sidebar, admin-only guard, admin theme) |
| `Pages/client/` | home, products, product details, cart, checkout, profile with tabs, favorites, orders, and `Sales/` (offers, bundles, my balance, waiting list) |
| `Pages/admin/` | dashboard, products, categories, orders, users, store settings (including shipping fees), domains, emails, and `Sales/` (coupons, offers, bundles, debts, waiting list) |
| `Pages/account/` | confirm email and reset password (opened from emails), `EmailLinks.jsx` (forwards email links), and `Invoice.jsx` (`/invoice/{id}`, printable, for customers and admins) |
| `Pages/platform/` | the platform console: login, stores, store form |
| `components/client`, `components/common` | shop building blocks: product cards, cart rows, coupon box, header, footer, search bar... |
| `components/admin/ui/` | the admin UI kit: Button, Card, Modal, ConfirmDialog, FormField, SelectInput, StatCard, StatusBadge, table classes... |
| `data/`, `constants/` | default texts (`storeContent.js`), settings defaults, tab and stage lists |
| `utils/` | `formatMoney`, `formatDate`, validation helpers |

## 16.2 The API layer

| File | Contents |
|---|---|
| `client.js` | `download(path)` saves a file the API generates (CSV exports) with the login token; `api(path, options)`: calls `/api` on the **same address**, adds the token, turns error bodies into an `ApiError` (`status`, `message`, `fields`, `retryAfter`), logs out on 401; `fetchAll` (every page of a long list); `upload` (FormData) |
| `mappers.js` | converts backend objects (snake_case) into the shapes the components use, and back |
| `auth.js`, `catalog.js`, `shop.js`, `admin.js`, `domains.js`, `platform.js`, `sales.js` | one function per endpoint, grouped by area |

**Why mappers?** The components were written with their own field names
(`nameAr`, `count`, `image`...). Rewriting every component to the backend's
names would have touched the whole UI. Instead, `mappers.js` translates in
one place, and when the backend changes a field only the mapper changes.

**Why the same address (`/api`) instead of a separate API domain?**

- The backend decides the store from the host name.
- There is no CORS to configure.
- One build serves every store and custom domain.

In development, `vite.config.js` passes `/api` and `/uploads` to the Go
server and **keeps the host name** (`changeOrigin: false`), so
`ahmad.localhost:5173` reaches store `ahmad`.

## 16.3 App-wide state (contexts)

| Provider | Holds |
|---|---|
| `AuthProvider` | the session (user, token, sign in/out), language, theme, toast messages, profile tabs, checkout form |
| `StoreProvider` | this store's settings: name, logo, colors (applied as CSS variables), page content |
| `CatalogProvider` | products and categories; `reload()` after admin changes or orders |
| `ShopProvider` | the signed-in customer's cart (`cart` = product lines; `cartInfo` = bundle lines, coupon, subtotal, discount, shipping, total), favorites and orders, plus every cart action; `setShippingCity(city)` makes every later cart refresh estimate shipping for that city (the checkout calls it when the customer picks a city) |

All cart actions go through `ShopProvider`:

- products: add, update, remove;
- bundles: add, update, remove;
- the coupon: apply, remove.

Each action calls the API and then reloads the cart from the backend, so
the screen always shows the server's prices and totals.

## 16.4 The screens added for sales features

**Customers**

| Screen | Where |
|---|---|
| Coupon box, bundle lines, backend totals | cart page and checkout summary |
| Pay Later | checkout → payment step, only when the store allows it |
| Notify me | product page, when sold out |
| Offers | `/client/offers` (tag icon in the top bar) |
| Bundles | `/client/bundles` (boxes icon in the top bar) |
| My Balance | profile, fifth tab (index 4 in `PROFILE_TABS`) |
| Waiting List | profile, sixth tab (index 5); Log Out moved to index 6 |

**Store admins**: Coupons, Offers, Bundles, Customer Debts and Waiting List
in the sidebar.

They share some pieces:

- `Pages/admin/Sales/shared.jsx`: form pieces such as the date range,
  "applies to" scope, product picker, status badge and on/off switch;
- `salesUtils.js`: the list loader and the add/edit modal state, which
  shows the backend's field errors under the right inputs.

## 16.5 Loading only what is needed

The admin panel and the platform console are loaded with `React.lazy`, only
when someone opens them. Shoppers no longer download them:

| | Before | After |
|---|---|---|
| main JavaScript file | 824 KB | 582 KB (182 KB gzipped) |
| admin panel | inside the main file | its own files, loaded on `/admin` |
| platform console | inside the main file | its own file, loaded on `admin.*` |

## 16.6 Language, theme and direction

- `useLang()` gives `tr("English", "عربي")` and `isAr`. Every text is
  written in both languages next to each other.
- Arabic switches the layout to right-to-left.
- The dark theme is a class plus CSS variables. The store's own colors
  replace the defaults through `useStoreTheme`.

## 16.7 Checks

```
cd Front-End
npm run lint     # ESLint, including React hook rules
npm run build    # production build into dist/
```
