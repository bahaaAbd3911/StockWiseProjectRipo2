-- 000006: categories as their own records (for the admin Categories page), and
-- the store settings the admin edits (name, logo, theme colors).

-- Products keep their category as text (products.category / category_ar), as
-- before. This table adds what the admin manages per category: the Arabic
-- name, an icon and an on/off switch. Renaming a category renames it on its
-- products too.
CREATE TABLE public.categories (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       character varying(100) NOT NULL,
    name_ar    character varying(100) DEFAULT '' NOT NULL,
    icon       character varying(100) DEFAULT '' NOT NULL, -- e.g. a Font Awesome class: fa-solid fa-laptop
    is_active  boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
CREATE UNIQUE INDEX categories_name_key ON public.categories (lower(name));

-- Every category already used by a product becomes a record.
INSERT INTO public.categories (name, name_ar)
SELECT DISTINCT ON (lower(category)) category, COALESCE(category_ar, '')
FROM public.products
WHERE category IS NOT NULL AND category <> ''
ORDER BY lower(category), category_ar NULLS LAST;

-- One row holding the store settings as JSON (see internal/models/settings.go).
CREATE TABLE public.store_settings (
    id         smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    data       jsonb DEFAULT '{}'::jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
INSERT INTO public.store_settings (id) VALUES (1);
