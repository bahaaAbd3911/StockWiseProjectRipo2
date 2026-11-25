import { useState } from "react";

import { Button, FormField, Modal, formGridClass } from "../../components/admin/ui";
import { useLang } from "../../hooks/useLang";

// The store's address is <slug>.<platform domain>; on admin.myplatform.com
// the platform domain is myplatform.com.
const platformDomain = () => window.location.hostname.replace(/^admin\./, "");

const SLUG_RE = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/;

// Create a store (with its admin) or edit one. The slug cannot change later:
// it is the store's address.
export default function StoreFormModal({ store, busy, onSave, onClose }) {
  const { tr } = useLang();
  const isNew = !store.id;
  const [form, setForm] = useState({
    slug: store.slug ?? "",
    name: store.name ?? "",
    currency: store.currency ?? "JOD",
    timezone: store.timezone ?? "Asia/Amman",
    firstName: "",
    lastName: "",
    email: "",
  });
  const [errors, setErrors] = useState({});
  const set = (key) => (e) => setForm((f) => ({ ...f, [key]: e.target.value }));

  const submit = (e) => {
    e.preventDefault();
    const found = {};
    if (isNew && !SLUG_RE.test(form.slug.trim()))
      found.slug = tr("Lower-case letters, digits and hyphens, e.g. ahmad-store", "أحرف إنجليزية صغيرة وأرقام وشرطات، مثل ahmad-store");
    if (!form.name.trim()) found.name = tr("Required", "مطلوب");
    if (isNew) {
      if (!form.firstName.trim()) found.firstName = tr("Required", "مطلوب");
      if (!form.lastName.trim()) found.lastName = tr("Required", "مطلوب");
      if (!/^\S+@\S+\.\S+$/.test(form.email.trim())) found.email = tr("Enter a valid email", "أدخل بريداً صحيحاً");
    }
    setErrors(found);
    if (Object.keys(found).length) return;

    const details = { name: form.name.trim(), currency: form.currency.trim(), timezone: form.timezone.trim() };
    onSave(
      isNew
        ? {
            ...details,
            slug: form.slug.trim(),
            admin: { first_name: form.firstName.trim(), last_name: form.lastName.trim(), email: form.email.trim() },
          }
        : details,
    );
  };

  return (
    <Modal
      title={isNew ? tr("New store", "متجر جديد") : tr(`Edit ${store.name}`, `تعديل ${store.name}`)}
      onClose={onClose}
      wide
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            {tr("Cancel", "إلغاء")}
          </Button>
          <Button type="submit" form="store-form" icon="fa-solid fa-check" disabled={busy}>
            {isNew ? tr("Create store", "إنشاء المتجر") : tr("Save", "حفظ")}
          </Button>
        </>
      }
    >
      <form id="store-form" onSubmit={submit} noValidate>
        {isNew && (
          <FormField
            label={tr("Address", "العنوان")}
            name="slug"
            value={form.slug}
            onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value.toLowerCase() }))}
            error={errors.slug}
            hint={`${form.slug || "store-name"}.${platformDomain()} · ${tr("cannot be changed later", "لا يمكن تغييره لاحقاً")}`}
          />
        )}
        <div className={formGridClass}>
          <FormField label={tr("Store name", "اسم المتجر")} name="name" value={form.name} onChange={set("name")} error={errors.name} />
          <FormField label={tr("Currency", "العملة")} name="currency" value={form.currency} onChange={set("currency")} hint="JOD, SAR, USD..." />
        </div>
        <FormField label={tr("Time zone", "المنطقة الزمنية")} name="timezone" value={form.timezone} onChange={set("timezone")} hint="Asia/Amman, Asia/Riyadh..." />

        {isNew && (
          <>
            <h3 style={{ margin: "16px 0 4px" }}>{tr("Store admin", "مدير المتجر")}</h3>
            <p style={{ margin: "0 0 8px", opacity: 0.75 }}>
              {tr(
                "They get an email with a link to choose their password, then manage the store at its address.",
                "سيصله بريد برابط لاختيار كلمة المرور، ثم يدير المتجر من عنوانه.",
              )}
            </p>
            <div className={formGridClass}>
              <FormField label={tr("First name", "الاسم الأول")} name="firstName" value={form.firstName} onChange={set("firstName")} error={errors.firstName} />
              <FormField label={tr("Last name", "اسم العائلة")} name="lastName" value={form.lastName} onChange={set("lastName")} error={errors.lastName} />
            </div>
            <FormField label={tr("Email", "البريد الإلكتروني")} name="email" type="email" value={form.email} onChange={set("email")} error={errors.email} />
          </>
        )}
      </form>
    </Modal>
  );
}
