-- Back to a single store: only safe while the database holds just store 1
-- (other stores' rows would be mixed together). Refuses otherwise.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.stores WHERE id <> 1) THEN
        RAISE EXCEPTION 'more than one store exists; export the other stores before going back';
    END IF;
END $$;

DO $$
DECLARE
    t text;
    r record;
BEGIN
    FOR t IN SELECT table_name FROM information_schema.columns
             WHERE table_schema = 'public' AND column_name = 'store_id' AND table_name <> 'store_domains'
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS store_isolation ON public.%I', t);
        EXECUTE format('ALTER TABLE public.%I DISABLE ROW LEVEL SECURITY', t);
    END LOOP;

    -- Store-aware foreign keys back to single-column ones.
    FOR r IN
        SELECT c.conname, child.relname AS child, parent.relname AS parent,
               a.attname AS col, c.confdeltype
        FROM pg_constraint c
        JOIN pg_class child ON child.oid = c.conrelid
        JOIN pg_class parent ON parent.oid = c.confrelid
        JOIN pg_namespace n ON n.oid = child.relnamespace AND n.nspname = 'public'
        JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = c.conkey[2]
        WHERE c.contype = 'f' AND array_length(c.conkey, 1) = 2 AND parent.relname <> 'stores'
    LOOP
        EXECUTE format('ALTER TABLE public.%I DROP CONSTRAINT %I', r.child, r.conname);
        EXECUTE format('ALTER TABLE public.%I ADD CONSTRAINT %I FOREIGN KEY (%I) REFERENCES public.%I (id) %s',
            r.child, r.conname, r.col, r.parent,
            CASE r.confdeltype WHEN 'c' THEN 'ON DELETE CASCADE' WHEN 'n' THEN 'ON DELETE SET NULL'
                               WHEN 'r' THEN 'ON DELETE RESTRICT' ELSE '' END);
    END LOOP;
END $$;

DROP INDEX IF EXISTS public.users_email_key;
ALTER TABLE public.users ADD CONSTRAINT users_email_key UNIQUE (email);
DROP INDEX IF EXISTS public.categories_name_key;
CREATE UNIQUE INDEX categories_name_key ON public.categories (lower(name));
DROP INDEX IF EXISTS public.coupons_code_key;
CREATE UNIQUE INDEX coupons_code_key ON public.coupons (upper(code));

ALTER TABLE public.store_settings DROP CONSTRAINT store_settings_pkey;
ALTER TABLE public.store_settings ADD COLUMN id smallint DEFAULT 1 NOT NULL CHECK (id = 1);
ALTER TABLE public.store_settings ADD PRIMARY KEY (id);

DO $$
DECLARE
    t text;
BEGIN
    FOR t IN SELECT table_name FROM information_schema.columns
             WHERE table_schema = 'public' AND column_name = 'store_id' AND table_name <> 'store_domains'
    LOOP
        EXECUTE format('ALTER TABLE public.%I DROP COLUMN store_id CASCADE', t);
    END LOOP;
END $$;

DROP TABLE IF EXISTS public.store_domains;
DROP TABLE IF EXISTS public.stores;
DROP FUNCTION IF EXISTS public.app_current_store();
DROP FUNCTION IF EXISTS public.app_is_platform();
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM stockwise_app;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM stockwise_app;
REVOKE USAGE ON SCHEMA public FROM stockwise_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON TABLES FROM stockwise_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON SEQUENCES FROM stockwise_app;
