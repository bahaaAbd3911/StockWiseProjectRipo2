import { FormField } from "../ui";
import styles from "./products.module.css";

export default function ProductMediaFields({ form, errors, update, tr }) {
  return (
    <>
      <div className={styles.mediaRow}>
        <FormField
          label={tr("Main image URL", "رابط الصورة الرئيسية")}
          name="image"
          type="url"
          placeholder="https://…"
          value={form.image}
          error={errors.image}
          onChange={(e) => update("image", e.target.value)}
        />
        {form.image && <img className={styles.preview} src={form.image} alt="" />}
      </div>

      <FormField
        as="textarea"
        label={tr("Gallery images", "صور المعرض")}
        name="images"
        hint={tr("One image URL per line", "رابط صورة واحد في كل سطر")}
        value={form.images.join("\n")}
        onChange={(e) => update("images", e.target.value.split("\n"))}
      />
    </>
  );
}
