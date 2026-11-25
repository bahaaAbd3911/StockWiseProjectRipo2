-- 000009: custom domains a store owner proves they control.
--
-- A store admin adds a domain (shop.example.com); it gets a random token. The
-- owner publishes it in DNS as a TXT record
--     _stockwise-verify.shop.example.com  TXT  "stockwise-verify=<token>"
-- and points the domain at the platform. When the API finds the record, the
-- domain is marked verified: only then is it used to find the store and only
-- then may an HTTPS certificate be issued for it.
ALTER TABLE public.store_domains ADD COLUMN verify_token character varying(64) DEFAULT '' NOT NULL;

-- At most one main (primary) domain per store: the address used in emails.
CREATE UNIQUE INDEX store_domains_one_primary ON public.store_domains (store_id) WHERE is_primary;
