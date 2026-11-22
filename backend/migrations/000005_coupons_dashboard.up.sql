-- 000005: coupon codes typed at checkout, and indexes for the admin dashboard.

-- A coupon takes a percentage or an amount off the order, ON TOP of the
-- automatic offers. It can apply to the whole order, to listed products, or
-- to one category.
CREATE TABLE public.coupons (
    id                 bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code               character varying(40) NOT NULL,            -- stored upper-case, matched case-insensitively
    description        text DEFAULT '' NOT NULL,
    description_ar     text DEFAULT '' NOT NULL,
    discount_type      character varying(20) NOT NULL,            -- percentage | fixed
    discount_value     numeric(10,2) NOT NULL,
    scope              character varying(20) DEFAULT 'all' NOT NULL, -- all | products | category
    category           character varying(100),
    min_order_amount   numeric(10,2) DEFAULT 0 NOT NULL,          -- cart subtotal needed to use it
    max_discount       numeric(10,2),                             -- cap for percentage coupons (NULL = no cap)
    starts_at          timestamp with time zone,                  -- NULL = valid from now
    ends_at            timestamp with time zone,                  -- NULL = no end
    usage_limit        integer,                                   -- total uses (NULL = unlimited)
    per_customer_limit integer DEFAULT 1,                         -- uses per customer (NULL = unlimited)
    is_active          boolean DEFAULT true NOT NULL,
    created_at         timestamp with time zone DEFAULT now() NOT NULL,
    updated_at         timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT coupons_discount_type_check CHECK (discount_type IN ('percentage', 'fixed')),
    CONSTRAINT coupons_discount_value_check CHECK (discount_value > 0 AND (discount_type <> 'percentage' OR discount_value <= 100)),
    CONSTRAINT coupons_scope_check CHECK (scope IN ('all', 'products', 'category')),
    CONSTRAINT coupons_category_check CHECK ((scope = 'category') = (category IS NOT NULL)),
    CONSTRAINT coupons_min_order_check CHECK (min_order_amount >= 0),
    CONSTRAINT coupons_max_discount_check CHECK (max_discount IS NULL OR max_discount > 0),
    CONSTRAINT coupons_window_check CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at),
    CONSTRAINT coupons_usage_limit_check CHECK (usage_limit IS NULL OR usage_limit > 0),
    CONSTRAINT coupons_per_customer_check CHECK (per_customer_limit IS NULL OR per_customer_limit > 0)
);
CREATE UNIQUE INDEX coupons_code_key ON public.coupons (upper(code));

-- Products a "products" coupon applies to.
CREATE TABLE public.coupon_products (
    coupon_id  bigint NOT NULL REFERENCES public.coupons(id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES public.products(id) ON DELETE CASCADE,
    PRIMARY KEY (coupon_id, product_id)
);
CREATE INDEX idx_coupon_products_product_id ON public.coupon_products (product_id);

-- One row per order that used a coupon; counts towards the usage limits.
-- Cancelling the order deletes the row, which gives the use back.
CREATE TABLE public.coupon_redemptions (
    id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    coupon_id       bigint NOT NULL REFERENCES public.coupons(id) ON DELETE CASCADE,
    user_id         bigint NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    order_id        bigint NOT NULL UNIQUE REFERENCES public.orders(id) ON DELETE CASCADE,
    discount_amount numeric(10,2) NOT NULL CHECK (discount_amount >= 0),
    created_at      timestamp with time zone DEFAULT now() NOT NULL
);
CREATE INDEX idx_coupon_redemptions_coupon_user ON public.coupon_redemptions (coupon_id, user_id);

-- The code a customer typed is kept on the cart until checkout.
ALTER TABLE public.carts
    ADD COLUMN coupon_id bigint REFERENCES public.coupons(id) ON DELETE SET NULL;

-- Orders: subtotal before the coupon, the discount, and a copy of the code
-- (kept even if the coupon is deleted later). total_amount = subtotal - discount.
ALTER TABLE public.orders
    ADD COLUMN subtotal_amount numeric(10,2),
    ADD COLUMN discount_amount numeric(10,2) DEFAULT 0 NOT NULL,
    ADD COLUMN coupon_id bigint REFERENCES public.coupons(id) ON DELETE SET NULL,
    ADD COLUMN coupon_code character varying(40);
UPDATE public.orders SET subtotal_amount = total_amount;
ALTER TABLE public.orders
    ALTER COLUMN subtotal_amount SET NOT NULL,
    ADD CONSTRAINT orders_discount_check CHECK (discount_amount >= 0 AND discount_amount <= subtotal_amount),
    ADD CONSTRAINT orders_total_check CHECK (total_amount = subtotal_amount - discount_amount);

-- Dashboard queries group orders by date and status.
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON public.orders (created_at);
CREATE INDEX IF NOT EXISTS idx_orders_status ON public.orders (status);
