import { useCallback, useEffect, useState } from "react";

import { formatMoney } from "../../../utils/format";
import { useToast } from "../../../hooks/useToast";

// Hooks and helpers shared by the Coupons, Offers, Bundles, Debts and Waiting list admin pages.

// Loads a list once, and again after every change (reload).
export function useAdminList(load) {
  const [items, setItems] = useState(null);
  const reload = useCallback(
    () => load().then(setItems, () => setItems([])),
    [load],
  );
  useEffect(() => {
    let current = true;
    load().then(
      (list) => current && setItems(list),
      () => current && setItems([]),
    );
    return () => {
      current = false;
    };
  }, [load]);
  return { items, reload };
}

// ISO date -> value for <input type="datetime-local"> (local time).
export const toLocalInput = (iso) => {
  if (!iso) return "";
  const d = new Date(iso);
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

export const discountLabel = (x) =>
  x.discountType === "percentage"
    ? `${x.discountValue}%`
    : formatMoney(x.discountValue);

// State for an add/edit modal: which item is open, the form, and the
// backend's field errors. save(fn) calls the API and closes on success.
export function useFormModal(toForm, onSaved) {
  const toast = useToast();
  const [open, setOpen] = useState(null); // null = closed, { id } = editing (id null = new)
  const [form, setForm] = useState(null);
  const [errors, setErrors] = useState(null);
  const [saving, setSaving] = useState(false);

  const show = (item) => {
    setOpen({ id: item?.id ?? null });
    setForm(toForm(item));
    setErrors(null);
  };
  const close = () => setOpen(null);
  const set = (key, value) => setForm((f) => ({ ...f, [key]: value }));

  const save = async (fn, success) => {
    setSaving(true);
    try {
      await fn(open.id, form);
      toast.success(success[0], success[1]);
      setOpen(null);
      await onSaved();
    } catch (err) {
      setErrors(err.fields ?? null);
      toast.error(err.message, err.message);
    } finally {
      setSaving(false);
    }
  };

  return { open, form, set, errors, saving, show, close, save };
}
