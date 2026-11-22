DROP TABLE IF EXISTS public.email_outbox;
DROP TABLE IF EXISTS public.email_verification_tokens;
ALTER TABLE public.users
    DROP COLUMN IF EXISTS debt_reminded_at,
    DROP COLUMN IF EXISTS email_verified_at;
