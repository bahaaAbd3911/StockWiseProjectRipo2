# 13. Emails

## 13.1 Every email

| Email | When | How it is sent |
|---|---|---|
| Confirm your email | sign-up, "send again" | directly, right after the response |
| Reset your password | forgot password | directly, right after the response |
| Store invitation | a platform admin creates a store or re-invites | directly |
| Order #N received | checkout | queue |
| Order #N: confirmed / processing / shipped / delivered / cancelled | every status change, customer cancel | queue |
| Payment received | admin records a payment | queue |
| Account balance updated | admin records an adjustment | queue |
| Friendly reminder: open balance | periodically while a balance is open | queue |
| X is back in stock | restock, or a request marked notified | queue |

Every email is in English then Arabic, and is **branded per store**:

- the store's name in the subject, the signature and the sender name
  (`"Omar Shop" <SMTP_FROM address>`);
- links to the store's own address (its main custom domain if it has one);
- amounts in the store's currency.

## 13.2 Where email links go

| Link in the email | Opens |
|---|---|
| `/verify-email?token=...` | the confirm-email page |
| `/reset-password?token=...` | the new-password page (also used by store invitations) |
| `/orders/{id}` | profile → My Orders |
| `/account/debt` | profile → My Balance |
| `/products/{id}` | the product page |

The last three are short, stable addresses that the React app forwards to
the right screen (`Pages/account/EmailLinks.jsx`). Emails already sent keep
working even if the app's internal routes change later.

## 13.3 The email queue

**Why a queue?** If the mail provider is slow or down, a checkout must not
hang or fail. So services only **store** the email with one fast `INSERT`
into `email_outbox`, and a background worker delivers it:

```
service ──INSERT──► email_outbox (sent_at NULL, attempts 0, next_attempt_at now)

notify.Worker every EMAIL_POLL_INTERVAL (5 s):
   SELECT due rows ... FOR UPDATE SKIP LOCKED LIMIT 20
   send each one
     ok    → sent_at = now
     error → attempts + 1, last_error, next_attempt_at = now + 1, 2, 4, 8 ... minutes (max 6 h)
   after 10 failed attempts the row stays "failed" for an admin to look at
```

What this design gives:

- **Emails survive restarts**, because they are rows, not memory.
- `FOR UPDATE SKIP LOCKED` lets several API servers share the queue without
  sending anything twice.
- **Retention:** delivered emails older than 30 days are deleted hourly. The
  queue is not an archive, and old emails contain customer data.

**Why are confirmation, reset and invite emails not queued?** The queue
stores the email's text, and their text contains the secret link. Storing it
would undo "only a hash of the token is stored". If the API restarts at that
moment, the user simply asks again.

## 13.4 Admin endpoints for the queue

| Endpoint | Use |
|---|---|
| `GET /api/admin/emails/summary` | `{pending, failed, sent_last_24h, oldest_pending_at}`; an old `oldest_pending_at` means delivery is stuck |
| `GET /api/admin/emails?status=pending\|failed\|sent&kind=&to=` | the list, newest first |
| `GET /api/admin/emails/{id}` | one email with its full text |
| `POST /api/admin/emails/{id}/retry` | send again now with fresh attempts |
| `POST /api/admin/emails/retry-failed` | the same for every failed email |
| `DELETE /api/admin/emails/{id}` | cancel an unsent email |

Sent emails cannot be retried or cancelled (409). There is no admin screen
for the queue yet ([25](25-limits-and-next-steps.md)).

## 13.5 Sending

| `MAIL_DRIVER` | What happens |
|---|---|
| `log` (development) | the email, links included, is printed in the API's log; nothing is sent |
| `smtp` (production) | sent through any provider with STARTTLS; headers are checked against injection |

For good delivery, send from your own domain with SPF and DKIM set up at the
provider.
