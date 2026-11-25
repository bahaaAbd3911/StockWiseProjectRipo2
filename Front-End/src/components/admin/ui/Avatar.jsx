import { initials } from "../../../utils/format";
import styles from "./ui.module.css";

const COLORS = ["#22c55e", "#3b82f6", "#f59e0b", "#8b5cf6", "#ef4444", "#14b8a6"];

export default function Avatar({ name, size = 40 }) {
  const color = COLORS[(name?.length ?? 0) % COLORS.length];

  return (
    <span
      className={styles.avatar}
      style={{ width: size, height: size, fontSize: size * 0.38, background: color }}
    >
      {initials(name)}
    </span>
  );
}
