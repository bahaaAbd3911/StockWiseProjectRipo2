import { CATEGORY_ICONS } from "../../../../constants/categoryIcons";
import styles from "../Categories.module.css";

export default function IconPicker({ value, onChange }) {
  return (
    <div className={styles.iconGrid} role="radiogroup">
      {CATEGORY_ICONS.map((icon) => (
        <button
          key={icon}
          type="button"
          role="radio"
          aria-checked={value === icon}
          aria-label={icon.replace("fa-solid fa-", "")}
          className={`${styles.iconOption} ${value === icon ? styles.iconSelected : ""}`}
          onClick={() => onChange(icon)}
        >
          <i className={icon} />
        </button>
      ))}
    </div>
  );
}
