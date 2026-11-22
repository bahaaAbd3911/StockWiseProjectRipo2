-- 000003: login revocation and password reset.

-- Every login token carries the user's token_version. Raising the number
-- (password change, password reset, "log out everywhere") makes every token
-- issued before it invalid at once.
ALTER TABLE public.users
    ADD COLUMN token_version integer DEFAULT 0 NOT NULL;

-- One row per "forgot password" request. Only the SHA-256 hash of the token is
-- stored: someone who reads this table still cannot use the links.
CREATE TABLE public.password_reset_tokens (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token_hash character(64) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    requested_ip character varying(45),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT password_reset_tokens_token_hash_key UNIQUE (token_hash)
);
CREATE INDEX idx_password_reset_tokens_user_id ON public.password_reset_tokens USING btree (user_id);
