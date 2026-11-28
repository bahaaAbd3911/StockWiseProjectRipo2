-- A private note per order for the store's admins (e.g. "customer asked for
-- delivery after 5 pm", "gift wrap"). Customers never see it: only the admin
-- order endpoints read these columns.
ALTER TABLE public.orders
    ADD COLUMN admin_note text NOT NULL DEFAULT '',
    ADD COLUMN admin_note_updated_at timestamp with time zone,
    ADD CONSTRAINT orders_admin_note_length CHECK (char_length(admin_note) <= 2000);
