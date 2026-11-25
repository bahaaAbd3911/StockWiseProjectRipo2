import styles from "./ui.module.css";

export default function IconButton({ icon, label, danger, onClick }) {
  return (
    <button
      type="button"
      title={label}
      aria-label={label}
      className={`${styles.iconBtn} ${danger ? styles.iconBtnDanger : ""}`}
      onClick={(e) => {
        // Buttons often live inside clickable rows/cards.
        e.stopPropagation();
        onClick?.(e);
      }}
    >
      <i className={icon} />
    </button>
  );
}
