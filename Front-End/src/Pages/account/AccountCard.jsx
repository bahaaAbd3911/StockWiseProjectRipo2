import { Link } from "react-router";

import { Card } from "../../components/admin/ui";
import { useLang } from "../../hooks/useLang";
import UserSettings from "../../components/common/UserSettings/UserSettings";
import styles from "./AccountPages.module.css";

// Shared frame of the pages opened from emails (confirm email, reset password).
export default function AccountCard({ icon, title, children }) {
  const { dir, tr } = useLang();

  return (
    <div className={styles.page} dir={dir}>
      <UserSettings />
      <Card className={styles.card}>
        <div className={styles.icon}>
          <i className={icon} />
        </div>
        <h1 className={styles.title}>{title}</h1>
        {children}
        <Link to="/client/home" className={styles.back}>
          {tr("Back to the store", "العودة إلى المتجر")}
        </Link>
      </Card>
    </div>
  );
}
