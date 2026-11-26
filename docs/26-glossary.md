# 26. Glossary

| Word | Meaning |
|---|---|
| tenant / store | one independent shop on the platform |
| multi-tenant | one installation serving many tenants, each seeing only its own data |
| slug | a store's short name in its address (`ahmad` in `ahmad.myplatform.com`) |
| platform domain | the domain the platform runs on (`myplatform.com`) |
| custom domain | a store's own domain name (`ahmadstore.com`) |
| platform admin | a person who runs the platform and creates stores |
| store admin | a person who manages one store |
| row-level security (RLS) | a PostgreSQL feature that filters rows per connection, below the application |
| platform mode | trusted server code allowed to see every store (email worker, platform API) |
| migration | a numbered SQL file that changes the database structure |
| transaction | several database changes that succeed or fail together |
| row lock (`FOR UPDATE`) | makes other requests wait for a row until the current change is finished |
| token (JWT) | the signed proof of login the browser sends with each request |
| token version | a number per user; raising it makes all older tokens invalid |
| rate limit / token bucket | a budget of requests that refills over time; over it → 429 |
| offer | an automatic, timed price cut |
| coupon | a code the customer types for a discount |
| bundle | several products sold as one cart line for one price |
| ledger | an append-only list of money movements; the balance is their sum |
| email queue / outbox | emails stored in the database and sent by a background worker |
| CSP | Content-Security-Policy: a header telling the browser which scripts, styles and images a page may load |
| HSTS | a header telling the browser to use only HTTPS for this site |
| clickjacking | another site showing your page in an invisible frame to trick clicks |
| on-demand TLS | getting an HTTPS certificate the first time a host is visited |
| CNAME / TXT | DNS record types: an alias to another name / a piece of text |
| SPF / DKIM | DNS records that prove an email really comes from your domain |
| distroless | a container image with only the program, no shell or tools |
| lazy loading | downloading a part of the app only when it is opened |
| mapper | a function converting the backend's data shape into the frontend's |
