import { isHexColor } from "../../../../utils/validators";
import styles from "../StoreSettings.module.css";

export default function ColorField({ label, hint, value, onChange }) {
  const valid = isHexColor(value);

  return (
    <div className={styles.colorField}>
      <h4>{label}</h4>
      <p>{hint}</p>
      <div className={`${styles.colorInput} ${valid ? "" : styles.invalid}`}>
        <input
          type="color"
          aria-label={label}
          value={valid && value.length === 7 ? value : "#000000"}
          onChange={(e) => onChange(e.target.value.toUpperCase())}
        />
        <input type="text" aria-label={`${label} hex`} value={value} maxLength={7} onChange={(e) => onChange(e.target.value)} />
      </div>
    </div>
  );
}
