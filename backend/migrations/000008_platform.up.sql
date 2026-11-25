-- 000008: the platform super admin, and per-store sender names on emails.

-- People who run the platform itself: create, edit, suspend and reactivate
-- stores. Deliberately not rows of users: a store login can never become a
-- platform login, and platform logins never belong to a store.
CREATE TABLE public.platform_admins (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         character varying(255) NOT NULL,
    first_name    character varying(100) DEFAULT '' NOT NULL,
    last_name     character varying(100) DEFAULT '' NOT NULL,
    password_hash character varying(255) NOT NULL,
    is_active     boolean DEFAULT true NOT NULL,
    token_version integer DEFAULT 0 NOT NULL,
    created_at    timestamp with time zone DEFAULT now() NOT NULL,
    updated_at    timestamp with time zone DEFAULT now() NOT NULL
);
CREATE UNIQUE INDEX platform_admins_email_key ON public.platform_admins (lower(email));
GRANT SELECT, INSERT, UPDATE, DELETE ON public.platform_admins TO stockwise_app;

-- The display name in From ("Ahmad Store" <no-reply@...>), so each store's
-- emails show the store's name.
ALTER TABLE public.email_outbox ADD COLUMN from_name character varying(150);
