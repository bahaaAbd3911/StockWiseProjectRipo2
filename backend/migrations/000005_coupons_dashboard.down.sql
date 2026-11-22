DROP INDEX IF EXISTS public.idx_orders_status;
DROP INDEX IF EXISTS public.idx_orders_created_at;
ALTER TABLE public.orders
    DROP CONSTRAINT IF EXISTS orders_total_check,
    DROP CONSTRAINT IF EXISTS orders_discount_check,
    DROP COLUMN IF EXISTS coupon_code,
    DROP COLUMN IF EXISTS coupon_id,
    DROP COLUMN IF EXISTS discount_amount,
    DROP COLUMN IF EXISTS subtotal_amount;
ALTER TABLE public.carts DROP COLUMN IF EXISTS coupon_id;
DROP TABLE IF EXISTS public.coupon_redemptions;
DROP TABLE IF EXISTS public.coupon_products;
DROP TABLE IF EXISTS public.coupons;
