-- Rollback of 000002. WARNING: deletes all offers, bundles and debt records.
-- Fails on purpose if any order still uses the new payment method/status.
ALTER TABLE public.orders DROP CONSTRAINT orders_payment_status_check;
ALTER TABLE public.orders ADD CONSTRAINT orders_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed'));
ALTER TABLE public.orders DROP CONSTRAINT check_payment_method;
ALTER TABLE public.orders ADD CONSTRAINT check_payment_method
    CHECK (payment_method = 'cash_on_delivery');

DROP TABLE IF EXISTS public.customer_debt_transactions;
ALTER TABLE public.users DROP COLUMN IF EXISTS credit_limit;

ALTER TABLE public.order_items DROP COLUMN IF EXISTS offer_name;
ALTER TABLE public.order_items DROP COLUMN IF EXISTS offer_id;

DROP TABLE IF EXISTS public.order_bundle_items;
DROP TABLE IF EXISTS public.order_bundles;
DROP TABLE IF EXISTS public.cart_bundles;
DROP TABLE IF EXISTS public.bundle_items;
DROP TABLE IF EXISTS public.bundles;
DROP TABLE IF EXISTS public.offer_products;
DROP TABLE IF EXISTS public.offers;
