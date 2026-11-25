import Sparkline from "./Sparkline";
import { toneColor } from "./tones";
import styles from "./ui.module.css";

const DEFAULT_TREND = [3, 5, 4, 7, 6, 9, 8, 11];

export default function StatCard({ icon, label, value, trend, down, tone = "green", spark }) {
  const color = toneColor(tone);

  return (
    <article className={`${styles.card} ${styles.statCard}`} style={{ "--tone": color }}>
      <span className={styles.statIcon}>
        <i className={icon} />
      </span>

      <div className={styles.statBody}>
        <p>{label}</p>
        <h3>{value}</h3>
        {trend && (
          <span className={`${styles.trend} ${down ? styles.trendDown : ""}`}>
            <i className={`fa-solid ${down ? "fa-arrow-down" : "fa-arrow-up"}`} /> {trend}
          </span>
        )}
      </div>

      <Sparkline values={spark ?? DEFAULT_TREND} color={color} />
    </article>
  );
}
