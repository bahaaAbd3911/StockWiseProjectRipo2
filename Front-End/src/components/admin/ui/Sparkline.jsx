import { useId } from "react";
import styles from "./ui.module.css";

const WIDTH = 96;
const HEIGHT = 44;

// Small decorative trend line used inside stat cards.
export default function Sparkline({ values, color }) {
  const gradientId = useId();
  const max = Math.max(...values);
  const min = Math.min(...values);
  const step = WIDTH / (values.length - 1);

  const points = values.map((v, i) => {
    const y = HEIGHT - 4 - ((v - min) / (max - min || 1)) * (HEIGHT - 8);
    return `${(i * step).toFixed(1)},${y.toFixed(1)}`;
  });

  return (
    <svg className={styles.sparkline} viewBox={`0 0 ${WIDTH} ${HEIGHT}`} aria-hidden>
      <defs>
        <linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.25" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      <polygon
        points={`0,${HEIGHT} ${points.join(" ")} ${WIDTH},${HEIGHT}`}
        fill={`url(#${gradientId})`}
      />
      <polyline points={points.join(" ")} fill="none" stroke={color} strokeWidth="2" />
    </svg>
  );
}
