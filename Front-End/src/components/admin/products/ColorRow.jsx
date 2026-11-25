import { IconButton, inputClass } from "../ui";
import styles from "./products.module.css";

export default function ColorRow({ color, onChange, onRemove, tr }) {
  const set = (key) => (e) => onChange({ ...color, [key]: e.target.value });

  return (
    <div className={styles.colorRow}>
      <input type="color" className={styles.swatch} value={color.hex} onChange={set("hex")} aria-label="Color" />
      <input className={inputClass} placeholder="#000000" value={color.hex} onChange={set("hex")} />
      <input className={inputClass} placeholder={tr("Name", "الاسم")} value={color.name} onChange={set("name")} />
      <input className={inputClass} dir="rtl" placeholder="الاسم بالعربية" value={color.nameAr} onChange={set("nameAr")} />
      <IconButton icon="fa-solid fa-trash" label={tr("Remove color", "حذف اللون")} danger onClick={onRemove} />
    </div>
  );
}
