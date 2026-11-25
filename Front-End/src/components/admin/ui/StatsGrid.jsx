import styles from "./ui.module.css";

export default function StatsGrid({ children }) {
  return <section className={styles.statsGrid}>{children}</section>;
}
