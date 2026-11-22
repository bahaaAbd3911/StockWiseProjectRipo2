-- Rollback of 000003. Pending password-reset links stop working.
DROP TABLE IF EXISTS public.password_reset_tokens;
ALTER TABLE public.users DROP COLUMN IF EXISTS token_version;
