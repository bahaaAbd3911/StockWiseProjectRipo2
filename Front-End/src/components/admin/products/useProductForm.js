import { useState } from "react";
import { EMPTY_PRODUCT, toProduct, validateProduct } from "./productForm";

export function useProductForm({ product, defaultCategory, categories, tr, onSave }) {
  const [form, setForm] = useState(() =>
    product ? { ...EMPTY_PRODUCT, ...product } : { ...EMPTY_PRODUCT, category: defaultCategory ?? "" },
  );
  const [errors, setErrors] = useState({});

  const update = (key, value) => setForm((prev) => ({ ...prev, [key]: value }));

  const submit = (e) => {
    e.preventDefault();
    const found = validateProduct(form, tr);
    setErrors(found);
    if (Object.keys(found).length === 0) onSave(toProduct(form, categories));
  };

  return { form, errors, update, submit };
}
