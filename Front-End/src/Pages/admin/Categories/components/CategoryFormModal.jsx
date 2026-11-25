import { useState } from "react";

import { Button, FormField, Modal, formGridClass } from "../../../../components/admin/ui";
import { CATEGORY_ICONS } from "../../../../constants/categoryIcons";
import IconPicker from "./IconPicker";
import styles from "../Categories.module.css";

function validate(form, categories, editingId, tr) {
  const errors = {};
  const taken = categories.some(
    (c) => c.id !== editingId && c.name.toLowerCase() === form.name.trim().toLowerCase(),
  );
  if (!form.name.trim()) errors.name = tr("Name is required", "الاسم مطلوب");
  else if (taken) errors.name = tr("This category already exists", "هذا التصنيف موجود بالفعل");
  if (!form.nameAr.trim()) errors.nameAr = tr("Arabic name is required", "الاسم العربي مطلوب");
  return errors;
}

export default function CategoryFormModal({ category, categories, tr, onSave, onClose }) {
  const [form, setForm] = useState({
    name: category?.name ?? "",
    nameAr: category?.nameAr ?? "",
    icon: category?.icon ?? CATEGORY_ICONS[0],
    isActive: category ? category.isActive !== false : true,
  });
  const [errors, setErrors] = useState({});
  const update = (key, value) => setForm((prev) => ({ ...prev, [key]: value }));

  const submit = (e) => {
    e.preventDefault();
    const found = validate(form, categories, category?.id, tr);
    setErrors(found);
    if (!Object.keys(found).length) onSave({ ...form, name: form.name.trim(), nameAr: form.nameAr.trim() });
  };

  return (
    <Modal
      title={category ? tr("Edit Category", "تعديل التصنيف") : tr("Add Category", "إضافة تصنيف")}
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>{tr("Cancel", "إلغاء")}</Button>
          <Button type="submit" form="categoryForm" icon="fa-solid fa-check">{tr("Save", "حفظ")}</Button>
        </>
      }
    >
      <form id="categoryForm" onSubmit={submit} noValidate>
        <div className={formGridClass}>
          <FormField label={tr("Category name", "اسم التصنيف")} name="name" value={form.name} error={errors.name} onChange={(e) => update("name", e.target.value)} />
          <FormField label={tr("Arabic name", "الاسم بالعربية")} name="nameAr" dir="rtl" value={form.nameAr} error={errors.nameAr} onChange={(e) => update("nameAr", e.target.value)} />
        </div>

        <FormField label={tr("Icon", "الأيقونة")}>
          <IconPicker value={form.icon} onChange={(icon) => update("icon", icon)} />
        </FormField>

        <label className={styles.toggle}>
          <input type="checkbox" checked={form.isActive} onChange={(e) => update("isActive", e.target.checked)} />
          {tr("Active (visible in the store)", "نشط (ظاهر في المتجر)")}
        </label>
      </form>
    </Modal>
  );
}
