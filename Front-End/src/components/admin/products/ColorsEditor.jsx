import { Button } from "../ui";
import ColorRow from "./ColorRow";
import { EMPTY_COLOR } from "./productForm";
import styles from "./products.module.css";

export default function ColorsEditor({ colors, error, onChange, tr }) {
  const replace = (index, color) => onChange(colors.map((c, i) => (i === index ? color : c)));
  const remove = (index) => onChange(colors.filter((_, i) => i !== index));

  return (
    <fieldset className={styles.colors}>
      <legend>{tr("Available colors", "الألوان المتوفرة")}</legend>

      {colors.map((color, index) => (
        <ColorRow
          key={index}
          color={color}
          tr={tr}
          onChange={(next) => replace(index, next)}
          onRemove={() => remove(index)}
        />
      ))}

      {error && <p className={styles.error}>{error}</p>}

      <Button variant="outline" icon="fa-solid fa-plus" onClick={() => onChange([...colors, { ...EMPTY_COLOR }])}>
        {tr("Add color", "إضافة لون")}
      </Button>
    </fieldset>
  );
}
