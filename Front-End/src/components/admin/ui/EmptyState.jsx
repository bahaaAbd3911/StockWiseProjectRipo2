import styles from "./ui.module.css";

export default function EmptyState({ icon = "fa-solid fa-box-open", text }) {
  return (
    <div className={styles.empty}>
      <i className={icon} />
      <p>{text}</p>
    </div>
  );
}
