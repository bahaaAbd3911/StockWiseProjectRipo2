-- Shipping fees. Each order keeps the fee it was charged (a copy, like the
-- prices and the address), and the total now includes it:
--   total_amount = subtotal_amount - discount_amount + shipping_amount
-- Existing orders get 0, which keeps them valid under the new rule.
-- The fee rules themselves live in store_settings.data -> 'shipping'.
ALTER TABLE public.orders
    ADD COLUMN shipping_amount numeric(10,2) NOT NULL DEFAULT 0,
    ADD CONSTRAINT orders_shipping_check CHECK (shipping_amount >= 0);

ALTER TABLE public.orders
    DROP CONSTRAINT orders_total_check,
    ADD CONSTRAINT orders_total_check CHECK (total_amount = subtotal_amount - discount_amount + shipping_amount);
