-- Baseline schema generated from the supplied PostgreSQL 18.6 schema-only dump.
CREATE TABLE public.cart_items (
    id bigint NOT NULL,
    cart_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer DEFAULT 1 NOT NULL,
    added_at timestamp with time zone DEFAULT now() NOT NULL,
    selected_color character varying(50),
    CONSTRAINT cart_items_quantity_check CHECK ((quantity > 0))
);
ALTER TABLE public.cart_items ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.cart_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.carts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.carts ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.carts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.favorites (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    product_id bigint NOT NULL,
    added_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.favorites ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.favorites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint,
    product_name character varying(255) NOT NULL,
    product_name_ar character varying(255),
    quantity integer NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    selected_color character varying(50),
    selected_color_ar character varying(50),
    subtotal numeric(10,2) NOT NULL,
    CONSTRAINT check_order_item_subtotal CHECK ((subtotal = ((quantity)::numeric * unit_price))),
    CONSTRAINT order_items_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT order_items_subtotal_check CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT order_items_unit_price_check CHECK ((unit_price >= (0)::numeric))
);
ALTER TABLE public.order_items ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.order_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    address_id bigint,
    recipient_name character varying(150) NOT NULL,
    phone character varying(20) NOT NULL,
    city character varying(100) NOT NULL,
    area character varying(100) NOT NULL,
    street character varying(255) NOT NULL,
    building_number character varying(50),
    apartment_number character varying(50),
    additional_instructions text,
    total_amount numeric(10,2) NOT NULL,
    status character varying(30) DEFAULT 'pending'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    payment_method character varying(30) DEFAULT 'cash_on_delivery'::character varying NOT NULL,
    payment_status character varying(30) DEFAULT 'pending'::character varying NOT NULL,
    CONSTRAINT check_payment_method CHECK (((payment_method)::text = 'cash_on_delivery'::text)),
    CONSTRAINT orders_payment_status_check CHECK (((payment_status)::text = ANY ((ARRAY['pending'::character varying, 'paid'::character varying, 'failed'::character varying])::text[]))),
    CONSTRAINT orders_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'confirmed'::character varying, 'processing'::character varying, 'shipped'::character varying, 'delivered'::character varying, 'cancelled'::character varying])::text[]))),
    CONSTRAINT orders_total_amount_check CHECK ((total_amount >= (0)::numeric))
);
ALTER TABLE public.orders ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.product_colors (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    name character varying(50) NOT NULL,
    name_ar character varying(50) NOT NULL,
    hex_code character varying(7) NOT NULL,
    CONSTRAINT product_colors_hex_code_check CHECK (((hex_code)::text ~ '^#[0-9A-Fa-f]{6}$'::text))
);
ALTER TABLE public.product_colors ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.product_colors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.product_images (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    image_url text NOT NULL,
    is_primary boolean DEFAULT false NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    CONSTRAINT product_images_display_order_check CHECK ((display_order >= 0))
);
ALTER TABLE public.product_images ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.product_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.product_reviews (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    user_id bigint NOT NULL,
    comment text NOT NULL,
    comment_ar text,
    rating integer NOT NULL,
    verified_purchase boolean DEFAULT false NOT NULL,
    helpful_count integer DEFAULT 0 NOT NULL,
    is_edited boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT product_reviews_helpful_count_check CHECK ((helpful_count >= 0)),
    CONSTRAINT product_reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);
ALTER TABLE public.product_reviews ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.product_reviews_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.products (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    name_ar character varying(255) NOT NULL,
    description text,
    description_ar text,
    category character varying(100),
    category_ar character varying(100),
    price numeric(10,2) NOT NULL,
    discount_value numeric(5,2) DEFAULT 0 NOT NULL,
    stock integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    CONSTRAINT products_discount_value_check CHECK (((discount_value >= (0)::numeric) AND (discount_value <= (100)::numeric))),
    CONSTRAINT products_price_check CHECK ((price >= (0)::numeric)),
    CONSTRAINT products_stock_check CHECK ((stock >= 0))
);
ALTER TABLE public.products ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.review_helpful_votes (
    id bigint NOT NULL,
    review_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.review_helpful_votes ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.review_helpful_votes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.user_addresses (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    address_label character varying(50),
    recipient_name character varying(150) NOT NULL,
    phone character varying(20) NOT NULL,
    city character varying(100) NOT NULL,
    area character varying(100) NOT NULL,
    street character varying(255) NOT NULL,
    building_number character varying(50),
    apartment_number character varying(50),
    additional_instructions text,
    is_default boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.user_addresses ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.user_addresses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.users (
    id bigint NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    phone character varying(20),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    profile_image_url text,
    is_active boolean DEFAULT true NOT NULL,
    role character varying(20) DEFAULT 'customer'::character varying NOT NULL
);
ALTER TABLE public.users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.waiting_list_items (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer DEFAULT 1 NOT NULL,
    selected_color character varying(50),
    status character varying(30) DEFAULT 'waiting'::character varying NOT NULL,
    requested_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    notified_at timestamp with time zone,
    CONSTRAINT waiting_list_items_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT waiting_list_items_status_check CHECK (((status)::text = ANY ((ARRAY['waiting'::character varying, 'notified'::character varying, 'fulfilled'::character varying, 'cancelled'::character varying])::text[])))
);
ALTER TABLE public.waiting_list_items ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.waiting_list_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_cart_id_product_id_key UNIQUE (cart_id, product_id);
ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_user_id_key UNIQUE (user_id);
ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_user_id_product_id_key UNIQUE (user_id, product_id);
ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.product_colors
    ADD CONSTRAINT product_colors_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.product_reviews
    ADD CONSTRAINT product_reviews_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.review_helpful_votes
    ADD CONSTRAINT review_helpful_votes_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.product_colors
    ADD CONSTRAINT unique_product_color UNIQUE (product_id, name);
ALTER TABLE ONLY public.review_helpful_votes
    ADD CONSTRAINT unique_review_user_vote UNIQUE (review_id, user_id);
ALTER TABLE ONLY public.product_reviews
    ADD CONSTRAINT unique_user_product_review UNIQUE (user_id, product_id);
ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT user_addresses_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.waiting_list_items
    ADD CONSTRAINT waiting_list_items_pkey PRIMARY KEY (id);
CREATE INDEX idx_cart_items_cart_id ON public.cart_items USING btree (cart_id);
CREATE INDEX idx_favorites_user_id ON public.favorites USING btree (user_id);
CREATE INDEX idx_order_items_order_id ON public.order_items USING btree (order_id);
CREATE INDEX idx_orders_user_id ON public.orders USING btree (user_id);
CREATE INDEX idx_product_reviews_product_id ON public.product_reviews USING btree (product_id);
CREATE INDEX idx_products_category ON public.products USING btree (category);
CREATE INDEX idx_user_addresses_user_id ON public.user_addresses USING btree (user_id);
CREATE INDEX idx_waiting_list_product_id ON public.waiting_list_items USING btree (product_id);
CREATE UNIQUE INDEX unique_default_address_per_user ON public.user_addresses USING btree (user_id) WHERE (is_default = true);
CREATE UNIQUE INDEX unique_primary_image_per_product ON public.product_images USING btree (product_id) WHERE (is_primary = true);
ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES public.carts(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.review_helpful_votes
    ADD CONSTRAINT fk_review_helpful_votes_review FOREIGN KEY (review_id) REFERENCES public.product_reviews(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.review_helpful_votes
    ADD CONSTRAINT fk_review_helpful_votes_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE SET NULL;
ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_address_id_fkey FOREIGN KEY (address_id) REFERENCES public.user_addresses(id) ON DELETE SET NULL;
ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE RESTRICT;
ALTER TABLE ONLY public.product_colors
    ADD CONSTRAINT product_colors_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.product_reviews
    ADD CONSTRAINT product_reviews_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.product_reviews
    ADD CONSTRAINT product_reviews_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT user_addresses_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.waiting_list_items
    ADD CONSTRAINT waiting_list_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.waiting_list_items
    ADD CONSTRAINT waiting_list_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
