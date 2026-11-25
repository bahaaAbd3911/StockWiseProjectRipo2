import { FaEye } from "react-icons/fa";
import { useNavigate } from "react-router";

import { formatDate, initials } from "../../../../utils/format";
import styles from "../Dashboard.module.css";

const AVATAR_COLORS = ["green", "pink", "blue"];

export default function UsersPanel({ t, users, isAr }) {
  const navigate = useNavigate();

  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3>{t.recentUsers}</h3>
        <button className={styles.viewAll} onClick={() => navigate("/admin/users")}>
          {t.viewAll} <FaEye />
        </button>
      </div>

      <div className={styles.userList}>
        {users.length === 0 && <p className={styles.panelSubtitle}>{t.noData}</p>}
        {users.map((user, i) => (
          <div
            className={styles.userItem}
            key={user.userID}
            onClick={() => navigate(`/admin/users/${user.userID}`)}
            style={{ cursor: "pointer" }}
          >
            <span className={`${styles.avatar} ${styles[AVATAR_COLORS[i % AVATAR_COLORS.length]]}`}>
              {initials(user.userFullName)}
            </span>
            <div className={styles.userInfo}>
              <strong>{user.userFullName}</strong>
              <small>{user.userEmail}</small>
            </div>
            <span className={styles.userJoined}>{formatDate(user.joinedAt, isAr)}</span>
          </div>
        ))}
      </div>
    </section>
  );
}
