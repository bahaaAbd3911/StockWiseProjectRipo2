DROP INDEX IF EXISTS public.store_domains_one_primary;
ALTER TABLE public.store_domains DROP COLUMN IF EXISTS verify_token;
