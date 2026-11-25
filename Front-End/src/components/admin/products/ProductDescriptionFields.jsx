import { FormField, formGridClass } from "../ui";

export default function ProductDescriptionFields({ form, update, tr }) {
  return (
    <div className={formGridClass}>
      <FormField
        as="textarea"
        label={tr("Description (English)", "الوصف بالإنجليزية")}
        name="description"
        value={form.description}
        onChange={(e) => update("description", e.target.value)}
      />
      <FormField
        as="textarea"
        dir="rtl"
        label={tr("Description (Arabic)", "الوصف بالعربية")}
        name="arabicDescription"
        value={form.arabicDescription}
        onChange={(e) => update("arabicDescription", e.target.value)}
      />
    </div>
  );
}
