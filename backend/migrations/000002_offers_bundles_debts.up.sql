-- 000002: offers (timed promotions), bundles, and customer debts (pay later).
-- Purely additive except for widening two CHECK constraints on orders.

------------------------------------------------------------------------------
-- OFFERS: timed promotions applied automatically to products or a category.
------------------------------------------------------------------------------
CREATE TABLE public.offers (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name character varying(150) NOT NULL,
    name_ar character varying(150) NOT NULL,
    description text,
    description_ar text,
    discount_type character varying(20) NOT NULL,
    discount_value numeric(10,2) NOT NULL,
    scope character varying(20) NOT NULL,
    category character varying(100),
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT offers_discount_type_check CHECK (discount_type IN ('percentage', 'fixed')),
    CONSTRAINT offers_discount_value_check CHECK (discount_value > 0),
    CONSTRAINT offers_percentage_max_check CHECK (discount_type <> 'percentage' OR discount_value <= 100),
    CONSTRAINT offers_scope_check CHECK (scope IN ('products', 'category')),
    CONSTRAINT offers_category_check CHECK ((scope = 'category') = (category IS NOT NULL)),
    CONSTRAINT offers_window_check CHECK (ends_at > starts_at)
);

-- Products covered by an offer whose scope is 'products'.
CREATE TABLE public.offer_products (
    offer_id bigint NOT NULL REFERENCES public.offers(id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES public.products(id) ON DELETE CASCADE,
    PRIMARY KEY (offer_id, product_id)
);

CREATE INDEX idx_offers_window ON public.offers USING btree (starts_at, ends_at) WHERE is_active;
CREATE INDEX idx_offer_products_product_id ON public.offer_products USING btree (product_id);

------------------------------------------------------------------------------
-- BUNDLES: several products sold together as one item at a bundle price.
------------------------------------------------------------------------------
CREATE TABLE public.bundles (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name character varying(255) NOT NULL,
    name_ar character varying(255) NOT NULL,
    description text,
    description_ar text,
    image_url text,
    price numeric(10,2) NOT NULL,
    starts_at timestamp with time zone,
    ends_at timestamp with time zone,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT bundles_price_check CHECK (price >= 0),
    CONSTRAINT bundles_window_check CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at)
);

-- The products inside a bundle. If a product has colors, the bundle fixes it.
CREATE TABLE public.bundle_items (
    bundle_id bigint NOT NULL REFERENCES public.bundles(id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES public.products(id) ON DELETE CASCADE,
    quantity integer NOT NULL,
    selected_color character varying(50),
    PRIMARY KEY (bundle_id, product_id),
    CONSTRAINT bundle_items_quantity_check CHECK (quantity > 0)
);
CREATE INDEX idx_bundle_items_product_id ON public.bundle_items USING btree (product_id);

-- Bundles placed in a cart (cart_items is reserved for single products).
CREATE TABLE public.cart_bundles (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cart_id bigint NOT NULL REFERENCES public.carts(id) ON DELETE CASCADE,
    bundle_id bigint NOT NULL REFERENCES public.bundles(id) ON DELETE CASCADE,
    quantity integer DEFAULT 1 NOT NULL,
    added_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT cart_bundles_quantity_check CHECK (quantity > 0),
    CONSTRAINT cart_bundles_cart_id_bundle_id_key UNIQUE (cart_id, bundle_id)
);

-- Snapshot of a bundle inside an order (names/prices copied at checkout).
CREATE TABLE public.order_bundles (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id bigint NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    bundle_id bigint REFERENCES public.bundles(id) ON DELETE SET NULL,
    bundle_name character varying(255) NOT NULL,
    bundle_name_ar character varying(255),
    quantity integer NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    subtotal numeric(10,2) NOT NULL,
    CONSTRAINT order_bundles_quantity_check CHECK (quantity > 0),
    CONSTRAINT order_bundles_unit_price_check CHECK (unit_price >= 0),
    CONSTRAINT order_bundles_subtotal_check CHECK (subtotal = quantity::numeric * unit_price)
);
CREATE INDEX idx_order_bundles_order_id ON public.order_bundles USING btree (order_id);

-- The products that were inside each ordered bundle (quantity = per bundle).
CREATE TABLE public.order_bundle_items (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_bundle_id bigint NOT NULL REFERENCES public.order_bundles(id) ON DELETE CASCADE,
    product_id bigint REFERENCES public.products(id) ON DELETE SET NULL,
    product_name character varying(255) NOT NULL,
    product_name_ar character varying(255),
    quantity integer NOT NULL,
    selected_color character varying(50),
    selected_color_ar character varying(50),
    CONSTRAINT order_bundle_items_quantity_check CHECK (quantity > 0)
);
CREATE INDEX idx_order_bundle_items_order_bundle_id ON public.order_bundle_items USING btree (order_bundle_id);

-- Which offer gave an order item its price (snapshot of the name too).
ALTER TABLE public.order_items
    ADD COLUMN offer_id bigint REFERENCES public.offers(id) ON DELETE SET NULL,
    ADD COLUMN offer_name character varying(150);

------------------------------------------------------------------------------
-- CUSTOMER DEBTS: credit limit per customer + an append-only ledger.
------------------------------------------------------------------------------
ALTER TABLE public.users
    ADD COLUMN credit_limit numeric(10,2) DEFAULT 0 NOT NULL,
    ADD CONSTRAINT users_credit_limit_check CHECK (credit_limit >= 0);

-- Every change to a customer's debt is one row. Balance = SUM(amount).
--   charge     (+) an order put on the customer's account
--   payment    (-) money received from the customer
--   reversal   (-) a charged order was cancelled
--   adjustment (+/-) manual correction by an admin (note required)
CREATE TABLE public.customer_debt_transactions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES public.users(id) ON DELETE RESTRICT,
    order_id bigint REFERENCES public.orders(id) ON DELETE SET NULL,
    type character varying(20) NOT NULL,
    amount numeric(10,2) NOT NULL,
    note text,
    created_by bigint REFERENCES public.users(id) ON DELETE SET NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT debt_type_check CHECK (type IN ('charge', 'payment', 'reversal', 'adjustment')),
    CONSTRAINT debt_amount_sign_check CHECK (
        (type = 'charge' AND amount > 0) OR
        (type IN ('payment', 'reversal') AND amount < 0) OR
        (type = 'adjustment' AND amount <> 0)),
    CONSTRAINT debt_adjustment_note_check CHECK (type <> 'adjustment' OR (note IS NOT NULL AND note <> ''))
);
CREATE INDEX idx_debt_transactions_user_id ON public.customer_debt_transactions USING btree (user_id, created_at);
CREATE INDEX idx_debt_transactions_order_id ON public.customer_debt_transactions USING btree (order_id);

-- Orders can now be paid later ("on account"), and a payment can be deferred.
ALTER TABLE public.orders DROP CONSTRAINT check_payment_method;
ALTER TABLE public.orders ADD CONSTRAINT check_payment_method
    CHECK (payment_method IN ('cash_on_delivery', 'pay_later'));

ALTER TABLE public.orders DROP CONSTRAINT orders_payment_status_check;
ALTER TABLE public.orders ADD CONSTRAINT orders_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'deferred'));
