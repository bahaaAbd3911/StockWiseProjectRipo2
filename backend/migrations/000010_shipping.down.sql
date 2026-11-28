-- Undo 000010. Refuses while orders with a shipping fee exist: their totals
-- would break the old rule (total = subtotal - discount).
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.orders WHERE shipping_amount <> 0) THEN
        RAISE EXCEPTION 'orders with a shipping fee exist; cannot remove shipping_amount';
    END IF;
END $$;

ALTER TABLE public.orders
    DROP CONSTRAINT orders_total_check,
    ADD CONSTRAINT orders_total_check CHECK (total_amount = subtotal_amount - discount_amount);

ALTER TABLE public.orders
    DROP CONSTRAINT orders_shipping_check,
    DROP COLUMN shipping_amount;
