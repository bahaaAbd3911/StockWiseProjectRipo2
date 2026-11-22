-- 000004: email verification at sign-up, and the email queue used for order,
-- debt and waiting-list notifications.

-- When the user proved they own their email address. NULL = not verified.
ALTER TABLE public.users
    ADD COLUMN email_verified_at timestamp with time zone,
    -- last time a "you still owe ..." reminder was queued for this customer
    ADD COLUMN debt_reminded_at timestamp with time zone;

-- Accounts that existed before verification was introduced are trusted, so
-- nobody already shopping is suddenly blocked at checkout.
UPDATE public.users SET email_verified_at = COALESCE(created_at, NOW());

-- Links sent to confirm an email address. Like password_reset_tokens, only the
-- SHA-256 hash of the token is stored.
CREATE TABLE public.email_verification_tokens (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    bigint NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token_hash character(64) NOT NULL UNIQUE,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
CREATE INDEX idx_email_verification_tokens_user_id ON public.email_verification_tokens (user_id);

-- Outgoing notification emails. Rows are written by the API and delivered by a
-- background worker, which retries failures with growing delays. Because the
-- queue is a table, emails survive restarts, and several API instances can
-- share it (FOR UPDATE SKIP LOCKED) without sending anything twice.
CREATE TABLE public.email_outbox (
    id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind            character varying(40) NOT NULL,  -- order_placed, debt_reminder, ...
    to_address      character varying(255) NOT NULL,
    subject         text NOT NULL,
    body            text NOT NULL,
    attempts        integer DEFAULT 0 NOT NULL,
    next_attempt_at timestamp with time zone DEFAULT now() NOT NULL,
    sent_at         timestamp with time zone,
    last_error      text,
    created_at      timestamp with time zone DEFAULT now() NOT NULL
);
-- The worker only ever looks for unsent rows that are due.
CREATE INDEX idx_email_outbox_due ON public.email_outbox (next_attempt_at) WHERE sent_at IS NULL;
