-- Undo 000011 (deletes every admin note).
ALTER TABLE public.orders
    DROP CONSTRAINT orders_admin_note_length,
    DROP COLUMN admin_note_updated_at,
    DROP COLUMN admin_note;
