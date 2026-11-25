import { salesSeries } from "../dashboardStats";
import { formatMoney } from "../../../../utils/format";
import styles from "../Dashboard.module.css";

const WIDTH = 495;
const TOP = 5;
const BOTTOM = 120;
const GRID = [5, 28, 51, 74, 97, 120];

// Revenue (non-cancelled orders) per day / week / month, drawn as an area chart.
export default function SalesChart({ t, period, setPeriod, orders, isAr }) {
  const series = salesSeries(orders, period, isAr);
  const total = series.reduce((sum, b) => sum + b.value, 0);
  const max = Math.max(...series.map((b) => b.value), 1);
  const step = WIDTH / Math.max(series.length - 1, 1);

  const points = series
    .map((b, i) => `${(i * step).toFixed(1)},${(BOTTOM - (b.value / max) * (BOTTOM - TOP)).toFixed(1)}`)
    .join(" ");
  const axis = GRID.map((_, i) => Math.round((max * (GRID.length - 1 - i)) / (GRID.length - 1)));

  return (
    <section className={`${styles.panel} ${styles.salesPanel}`}>
      <div className={styles.panelHeader}>
        <div>
          <h3>{t.salesOverview}</h3>
          <p className={styles.panelSubtitle}>{formatMoney(total)}</p>
        </div>

        <div className={styles.periodButtons}>
          {["daily", "weekly", "monthly"].map((item) => (
            <button
              key={item}
              className={period === item ? styles.selectedPeriod : ""}
              onClick={() => setPeriod(item)}
            >
              {t[item]}
            </button>
          ))}
        </div>
      </div>

      <div className={styles.chart}>
        <div className={styles.chartValues}>
          {axis.map((value, i) => (
            <span key={i}>{value}</span>
          ))}
        </div>

        <svg viewBox="0 0 500 125" preserveAspectRatio="none" role="img" aria-label={t.salesOverview}>
          {GRID.map((y) => (
            <line key={y} x1="0" y1={y} x2="500" y2={y} className={styles.gridLine} />
          ))}
          <polygon points={`0,125 ${points} ${WIDTH},125`} className={styles.chartArea} />
          <polyline points={points} className={styles.chartLine} />
        </svg>
      </div>

      <div className={styles.chartLabels}>
        {series.map((b) => (
          <span key={b.label}>{b.label}</span>
        ))}
      </div>
    </section>
  );
}
