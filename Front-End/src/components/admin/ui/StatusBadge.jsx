import { toneColor } from "./tones";
import styles from "./ui.module.css";

export default function StatusBadge({ tone = "green", icon, children }) {
  return (
    <span className={styles.badge} style={{ "--tone": toneColor(tone) }}>
      {icon ? <i className={icon} /> : <span className={styles.badgeDot} />}
      {children}
    </span>
  );
}
