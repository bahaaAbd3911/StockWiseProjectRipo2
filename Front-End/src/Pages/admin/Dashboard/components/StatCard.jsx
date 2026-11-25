import { FaArrowUp, FaArrowDown } from "react-icons/fa";

import styles from "../Dashboard.module.css";

export default function StatCard({ item, t }) {
  const Icon = item.icon;

  return (
    <article className={styles.statCard}>
      <div className={styles.statTop}>
        <span className={styles.statIcon}>
          <Icon />
        </span>

        <span className={styles.statChange}>
          {item.positive ? <FaArrowUp /> : <FaArrowDown />}
          {item.change}
        </span>
      </div>

      <p className={styles.statLabel}>{t[item.key]}</p>

      <h2>{item.value}</h2>

      <span className={styles.statHint}>{t.fromLastMonth}</span>
    </article>
  );
}
