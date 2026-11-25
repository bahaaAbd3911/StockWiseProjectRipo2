import { Card } from "../../../../components/admin/ui";
import styles from "../StoreSettings.module.css";

export default function SettingsCard({ icon, title, subtitle, children }) {
  return (
    <Card className={styles.card}>
      <div className={styles.cardHead}>
        <i className={icon} />
        <div>
          <h3>{title}</h3>
          {subtitle && <p>{subtitle}</p>}
        </div>
      </div>
      {children}
    </Card>
  );
}
