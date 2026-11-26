# 25. Limits and next steps

## Before going live (your accounts are needed)

These cannot be done from the code. They need your server, domain and email
accounts. The steps are in [21-deployment-and-operations.md](21-deployment-and-operations.md).

- [ ] Rent the server and install Docker.
- [ ] Point DNS at it (`myplatform.com` and `*.myplatform.com`).
- [ ] Set up an SMTP provider with SPF and DKIM.
- [ ] Fill in `.env.production`.
- [ ] Configure backups off the server (`BACKUP_REMOTE`) and a monitoring
      check.
- [ ] Run `govulncheck` and `staticcheck` once ([23](23-security-and-quality-review.md)).
- [ ] Open the shop in a browser and check the developer console for
      Content-Security-Policy errors.

## Worth doing soon

| Item | Why |
|---|---|
| An admin "Emails" page | the email-queue endpoints exist but have no screen |
| `npm audit fix` | development-tool warnings |
| Browser tests (e.g. Playwright) for the main flows | the API side of every screen is tested, the clicking is not |

## When the platform grows

| Item | When |
|---|---|
| Wildcard certificate for store subdomains | with many stores (Let's Encrypt: 50 certificates / week / domain) |
| Rate-limit counters in Redis | with more than one API server |
| Image resizing / a CDN | with many product photos |
| Login in http-only cookies | if the risk of the token in localStorage becomes unacceptable (needs CSRF protection) |

## Features not built

- Online card payments (only cash on delivery and pay later).
- SMS notifications.
- CAPTCHA on sign-up and login.
- Paying off a specific order: debt is tracked per customer, so an order's
  payment status stays `deferred` after the customer pays.
- Credit in the customer's favour: payments larger than the debt are
  refused.
- Offers on bundles: a bundle's price is already the deal.
