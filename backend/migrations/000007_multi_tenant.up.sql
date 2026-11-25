-- 000007: one database, many stores.
--
-- Every row of a store's data carries store_id. Isolation is enforced by the
-- database itself, in three layers:
--
--   1. Row-level security (RLS): the API runs as the role stockwise_app, and
--      on every table a policy limits reads AND writes to the rows of the
--      current store. The current store is a per-connection setting
--      (app.store_id) the API sets from the request's host name. With no
--      store set, nothing is visible: forgetting it fails closed.
--   2. store_id defaults to that setting, so inserts land in the current
--      store without the code passing it (and the policy refuses any other).
--   3. Foreign keys include store_id ((store_id, product_id) -> products
--      (store_id, id)), so a row can never point at another store's row,
--      even where RLS does not apply (foreign key checks ignore it).
--
-- Existing data becomes the first store, "main".

-- ---------- stores ----------

CREATE TABLE public.stores (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug       character varying(63) NOT NULL,   -- the subdomain: slug.PLATFORM_DOMAIN
    name       character varying(150) NOT NULL,
    status     character varying(20) DEFAULT 'active' NOT NULL,
    currency   character varying(10) DEFAULT '' NOT NULL,
    timezone   character varying(64) DEFAULT 'Asia/Amman' NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT stores_status_check CHECK (status IN ('active', 'suspended')),
    -- a valid DNS label: lower-case letters, digits and inner hyphens
    CONSTRAINT stores_slug_check CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$')
);
CREATE UNIQUE INDEX stores_slug_key ON public.stores (slug);

-- Extra host names of a store (custom domains, e.g. ahmadstore.com).
CREATE TABLE public.store_domains (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    store_id    bigint NOT NULL REFERENCES public.stores(id) ON DELETE CASCADE,
    domain      character varying(253) NOT NULL,
    is_primary  boolean DEFAULT false NOT NULL,
    verified_at timestamp with time zone,         -- set once ownership is proven (step 4)
    created_at  timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT store_domains_domain_check CHECK (domain = lower(domain))
);
CREATE UNIQUE INDEX store_domains_domain_key ON public.store_domains (domain);
CREATE INDEX idx_store_domains_store_id ON public.store_domains (store_id);

-- Store 1 holds everything that existed before; new stores get 2, 3, ...
INSERT INTO public.stores (id, slug, name) OVERRIDING SYSTEM VALUE VALUES (1, 'main', 'StockWise');
SELECT setval(pg_get_serial_sequence('public.stores', 'id'), 1);

-- ---------- the current store, as seen by SQL ----------

-- Set per connection by the API: SELECT set_config('app.store_id', '42', false)
CREATE FUNCTION public.app_current_store() RETURNS bigint
    LANGUAGE sql STABLE
    AS $$ SELECT NULLIF(current_setting('app.store_id', true), '')::bigint $$;

-- Platform mode (background jobs that work across stores, the platform admin):
-- set only by trusted server code, never from a request value.
CREATE FUNCTION public.app_is_platform() RETURNS boolean
    LANGUAGE sql STABLE
    AS $$ SELECT COALESCE(current_setting('app.platform', true), '') = 'on' $$;

-- ---------- store_id on every store table ----------

DO $$
DECLARE
    t text;
    tenant_tables text[] := ARRAY[
        'users', 'user_addresses', 'carts', 'cart_items', 'cart_bundles', 'favorites',
        'products', 'product_images', 'product_colors', 'product_reviews', 'review_helpful_votes',
        'categories', 'offers', 'offer_products', 'bundles', 'bundle_items',
        'coupons', 'coupon_products', 'coupon_redemptions',
        'orders', 'order_items', 'order_bundles', 'order_bundle_items',
        'waiting_list_items', 'customer_debt_transactions',
        'password_reset_tokens', 'email_verification_tokens', 'email_outbox', 'store_settings'];
BEGIN
    FOREACH t IN ARRAY tenant_tables LOOP
        EXECUTE format('ALTER TABLE public.%I ADD COLUMN store_id bigint', t);
        EXECUTE format('UPDATE public.%I SET store_id = 1', t);
        EXECUTE format('ALTER TABLE public.%I ALTER COLUMN store_id SET DEFAULT public.app_current_store()', t);
        EXECUTE format('ALTER TABLE public.%I ALTER COLUMN store_id SET NOT NULL', t);
        EXECUTE format('ALTER TABLE public.%I ADD CONSTRAINT %I FOREIGN KEY (store_id) REFERENCES public.stores(id) ON DELETE CASCADE',
                       t, t || '_store_id_fkey');
        EXECUTE format('CREATE INDEX %I ON public.%I (store_id)', 'idx_' || t || '_store_id', t);
        -- (store_id, id) is the target of the store-aware foreign keys below.
        IF EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_schema = 'public' AND table_name = t AND column_name = 'id') THEN
            EXECUTE format('ALTER TABLE public.%I ADD CONSTRAINT %I UNIQUE (store_id, id)', t, t || '_store_id_id_key');
        END IF;
    END LOOP;

    -- Every foreign key between two store tables also has to match store_id:
    -- (store_id, x_id) -> parent (store_id, id). Rebuilt from the catalog so
    -- none is missed. ON DELETE behaviour is kept; SET NULL only clears the
    -- reference column, never store_id.
    DECLARE
        r record;
        action text;
    BEGIN
        FOR r IN
            SELECT c.conname, child.relname AS child, parent.relname AS parent,
                   a.attname AS col, c.confdeltype
            FROM pg_constraint c
            JOIN pg_class child ON child.oid = c.conrelid
            JOIN pg_class parent ON parent.oid = c.confrelid
            JOIN pg_namespace n ON n.oid = child.relnamespace AND n.nspname = 'public'
            JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = c.conkey[1]
            WHERE c.contype = 'f'
              AND array_length(c.conkey, 1) = 1
              AND child.relname = ANY (tenant_tables)
              AND parent.relname = ANY (tenant_tables)
        LOOP
            action := CASE r.confdeltype
                WHEN 'c' THEN 'ON DELETE CASCADE'
                WHEN 'n' THEN format('ON DELETE SET NULL (%I)', r.col)
                WHEN 'r' THEN 'ON DELETE RESTRICT'
                ELSE 'ON DELETE NO ACTION' END;
            EXECUTE format('ALTER TABLE public.%I DROP CONSTRAINT %I', r.child, r.conname);
            EXECUTE format('ALTER TABLE public.%I ADD CONSTRAINT %I FOREIGN KEY (store_id, %I) REFERENCES public.%I (store_id, id) %s',
                           r.child, r.conname, r.col, r.parent, action);
        END LOOP;
    END;
END $$;

-- ---------- uniqueness per store ----------

-- The same person can have an account in two stores.
ALTER TABLE public.users DROP CONSTRAINT users_email_key;
CREATE UNIQUE INDEX users_email_key ON public.users (store_id, email);

DROP INDEX public.categories_name_key;
CREATE UNIQUE INDEX categories_name_key ON public.categories (store_id, lower(name));

DROP INDEX public.coupons_code_key;
CREATE UNIQUE INDEX coupons_code_key ON public.coupons (store_id, upper(code));

-- One settings row per store (was: one row, id = 1).
ALTER TABLE public.store_settings DROP CONSTRAINT store_settings_pkey;
ALTER TABLE public.store_settings DROP CONSTRAINT IF EXISTS store_settings_id_check;
ALTER TABLE public.store_settings DROP COLUMN id;
ALTER TABLE public.store_settings ADD PRIMARY KEY (store_id);

-- ---------- row-level security ----------

DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT c.table_name FROM information_schema.columns c
        JOIN information_schema.tables tb ON tb.table_schema = c.table_schema AND tb.table_name = c.table_name
        WHERE c.table_schema = 'public' AND c.column_name = 'store_id'
          AND tb.table_type = 'BASE TABLE' AND c.table_name <> 'store_domains'
    LOOP
        EXECUTE format('ALTER TABLE public.%I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format($p$CREATE POLICY store_isolation ON public.%I
            USING (public.app_is_platform() OR store_id = public.app_current_store())
            WITH CHECK (public.app_is_platform() OR store_id = public.app_current_store())$p$, t);
    END LOOP;
END $$;

-- ---------- the role the API runs as ----------

-- Roles belong to the whole PostgreSQL server, so it may exist already (for
-- example from another database on the same server).
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'stockwise_app') THEN
        CREATE ROLE stockwise_app NOLOGIN;
    END IF;
    -- The login user must be allowed to SET ROLE stockwise_app.
    BEGIN
        EXECUTE format('GRANT stockwise_app TO %I', current_user);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'could not grant stockwise_app to %: %', current_user, SQLERRM;
    END;
END $$;

GRANT USAGE ON SCHEMA public TO stockwise_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO stockwise_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO stockwise_app;
REVOKE ALL ON public.schema_migrations FROM stockwise_app;
-- Tables added by later migrations get the same rights.
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO stockwise_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO stockwise_app;
