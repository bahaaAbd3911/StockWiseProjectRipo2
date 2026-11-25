import { useNavigate } from "react-router";

import { useLang } from "../../../hooks/useLang";
import styles from "./AccessNotice.module.css";

// Full-page message for blocked or loading routes.
export default function AccessNotice({ icon, title, text, actionLabel, actionTo, loading }) {
  const navigate = useNavigate();
  const { dir, tr } = useLang();

  return (
    <div className={styles.page} dir={dir}>
      <div className={styles.card} role={loading ? "status" : "alert"}>
        <i className={`${loading ? "fa-solid fa-spinner fa-spin" : icon} ${styles.icon}`} />
        <h1>{title}</h1>
        {text && <p>{text}</p>}
        {!loading && (
          <div className={styles.actions}>
            {actionLabel && (
              <button className={styles.primary} onClick={() => navigate(actionTo)}>
                {actionLabel}
              </button>
            )}
            <button className={styles.secondary} onClick={() => navigate("/client/home")}>
              {tr("Back to store", "العودة إلى المتجر")}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
