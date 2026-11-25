import { FaClock } from "react-icons/fa";

import { ACTIVITY_ICONS } from "../dashboardData";
import { timeAgo } from "../dashboardStats";
import styles from "../Dashboard.module.css";

export default function ActivityPanel({ t, activities }) {
  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3>{t.recentActivity}</h3>
      </div>

      <div className={styles.activityList}>
        {activities.length === 0 && <p className={styles.panelSubtitle}>{t.noData}</p>}
        {activities.map((item, index) => {
          const Icon = ACTIVITY_ICONS[item.action];
          return (
            <div className={styles.activityItem} key={index}>
              <span className={styles.activityIcon}>
                <Icon />
              </span>
              <div>
                <p>
                  <strong>{item.name}</strong> {t[item.action]}
                </p>
                <small>
                  <FaClock /> {timeAgo(item.date, t)}
                </small>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
