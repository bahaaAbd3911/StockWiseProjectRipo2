import { useState } from "react";

import {
  FormField,
  SelectInput,
  StatusBadge,
  inputClass,
} from "../../../components/admin/ui";
import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import styles from "./Sales.module.css";

// Form pieces shared by the Coupons, Offers, Bundles, Debts and Waiting list admin pages.

const STATUS = {
  live: ["green", "Live", "فعّال"],
  scheduled: ["blue", "Scheduled", "مجدول"],
  expired: ["gray", "Expired", "منتهي"],
  disabled: ["orange", "Off", "متوقف"],
  waiting: ["orange", "Waiting", "بالانتظار"],
  notified: ["blue", "Notified", "تم الإبلاغ"],
  fulfilled: ["green", "Bought", "تم الشراء"],
  cancelled: ["gray", "Cancelled", "ملغى"],
};

export function SaleStatus({ status }) {
  const { tr } = useLang();
  const [tone, en, ar] = STATUS[status] ?? ["gray", status, status];
  return <StatusBadge tone={tone}>{tr(en, ar)}</StatusBadge>;
}

// Field errors from the backend come back as { field: message }.
const fieldError = (errors, key) => errors?.[key];

// required: offers must have both dates; coupons and bundles may leave them empty.
export function DateRangeFields({ form, set, errors, required = false }) {
  const { tr } = useLang();
  const opt = (en, ar) =>
    required ? tr(en, ar) : tr(`${en} (optional)`, `${ar} (اختياري)`);
  return (
    <>
      <FormField
        label={opt("Starts", "يبدأ")}
        required={required}
        name="startsAt"
        type="datetime-local"
        value={form.startsAt}
        onChange={(e) => set("startsAt", e.target.value)}
        error={fieldError(errors, "starts_at")}
      />
      <FormField
        label={opt("Ends", "ينتهي")}
        required={required}
        name="endsAt"
        type="datetime-local"
        value={form.endsAt}
        onChange={(e) => set("endsAt", e.target.value)}
        error={fieldError(errors, "ends_at")}
      />
    </>
  );
}

export function ActiveToggle({ form, set }) {
  const { tr } = useLang();
  return (
    <label className={styles.check}>
      <input
        type="checkbox"
        checked={form.isActive}
        onChange={(e) => set("isActive", e.target.checked)}
      />
      {tr("Switched on", "مفعّل")}
    </label>
  );
}

// "What does it apply to": the whole order (coupons only), some products, or one category.
export function ScopeFields({ form, set, errors, allowAll }) {
  const { tr, isAr } = useLang();
  const { categories } = useCatalog();
  const scopes = [
    ...(allowAll
      ? [{ value: "all", label: tr("The whole order", "الطلب كاملاً") }]
      : []),
    { value: "products", label: tr("Chosen products", "منتجات محددة") },
    { value: "category", label: tr("One category", "تصنيف واحد") },
  ];

  return (
    <>
      <FormField
        label={tr("Applies to", "ينطبق على")}
        error={fieldError(errors, "scope")}
      >
        <SelectInput
          value={form.scope}
          onChange={(v) => set("scope", v)}
          options={scopes}
        />
      </FormField>
      {form.scope === "category" && (
        <FormField
          label={tr("Category", "التصنيف")}
          error={fieldError(errors, "category")}
        >
          <SelectInput
            value={form.category}
            onChange={(v) => set("category", v)}
            options={[
              { value: "", label: tr("Choose…", "اختر…") },
              ...categories.map((c) => ({
                value: c.name,
                label: isAr ? c.nameAr || c.name : c.name,
              })),
            ]}
          />
        </FormField>
      )}
      {form.scope === "products" && (
        <div className={styles.full}>
          <ProductPicker
            selected={form.productIds}
            onChange={(ids) => set("productIds", ids)}
            error={fieldError(errors, "product_ids")}
          />
        </div>
      )}
    </>
  );
}

// Searchable list of the store's products with checkboxes.
export function ProductPicker({ selected, onChange, error }) {
  const { tr, isAr } = useLang();
  const { products } = useCatalog();
  const [q, setQ] = useState("");
  const query = q.trim().toLowerCase();
  const shown = products.filter(
    (p) =>
      !query ||
      p.name.toLowerCase().includes(query) ||
      (p.nameAr ?? "").includes(q.trim()),
  );
  const toggle = (id) =>
    onChange(
      selected.includes(id)
        ? selected.filter((x) => x !== id)
        : [...selected, id],
    );

  return (
    <FormField
      label={tr(
        `Products (${selected.length} chosen)`,
        `المنتجات (${selected.length} محدد)`,
      )}
      error={error}
    >
      <div className={styles.picker}>
        <input
          className={inputClass}
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder={tr("Search products", "ابحث عن منتج")}
        />
        <div className={styles.pickerList}>
          {shown.map((p) => (
            <label key={p.id} className={styles.check}>
              <input
                type="checkbox"
                checked={selected.includes(p.id)}
                onChange={() => toggle(p.id)}
              />
              {isAr ? p.nameAr || p.name : p.name}
              <small className={styles.muted}>#{p.id}</small>
            </label>
          ))}
          {shown.length === 0 && (
            <small className={styles.muted}>
              {tr("No products found", "لا توجد منتجات")}
            </small>
          )}
        </div>
      </div>
    </FormField>
  );
}
