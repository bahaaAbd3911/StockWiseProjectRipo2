import { FormField, formGridClass, selectClass } from "../ui";

export default function ProductNameFields({ form, errors, update, categories, lockCategory, tr, isAr }) {
  return (
    <div className={formGridClass}>
      <FormField
        label={tr("Name (English)", "الاسم بالإنجليزية")}
        name="name"
        value={form.name}
        error={errors.name}
        onChange={(e) => update("name", e.target.value)}
      />
      <FormField
        label={tr("Name (Arabic)", "الاسم بالعربية")}
        name="nameAr"
        dir="rtl"
        value={form.nameAr}
        error={errors.nameAr}
        onChange={(e) => update("nameAr", e.target.value)}
      />
      <FormField label={tr("Category", "التصنيف")} error={errors.category}>
        <select
          className={selectClass}
          value={form.category}
          disabled={lockCategory}
          onChange={(e) => update("category", e.target.value)}
        >
          <option value="">{tr("Select category", "اختر التصنيف")}</option>
          {categories.map((c) => (
            <option key={c.id} value={c.name}>
              {isAr ? c.nameAr : c.name}
            </option>
          ))}
        </select>
      </FormField>
    </div>
  );
}
