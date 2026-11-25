import styles from "./ui.module.css";

export default function PageHeader({ icon, title, subtitle, children }) {
  return (
    <header className={styles.pageHeader}>
      <div className={styles.headerTitle}>
        <i className={`${icon} ${styles.headerIcon}`} />
        <div>
          <h1>{title}</h1>
          {subtitle && <p>{subtitle}</p>}
        </div>
      </div>

      {children && <div className={styles.headerActions}>{children}</div>}
    </header>
  );
}
