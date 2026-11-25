import { Avatar, Button, Card } from "../../../../components/admin/ui";
import { formatDate } from "../../../../utils/format";
import { RoleBadge, UserStatusBadge } from "../../Users/components/UserBadges";
import styles from "../UserProfile.module.css";

export default function ProfileCard({ user, isAr, tr, onEdit, onDelete }) {
  const details = [
    ["fa-solid fa-id-badge", `#${user.userID}`],
    ["fa-solid fa-envelope", user.userEmail],
    ["fa-solid fa-phone", user.phone || "—"],
    ["fa-solid fa-calendar", tr(`Joined ${formatDate(user.joinedAt, isAr)}`, `انضم ${formatDate(user.joinedAt, isAr)}`)],
  ];

  return (
    <Card className={styles.profile}>
      <Avatar name={user.userFullName} size={92} />

      <div className={styles.identity}>
        <h1>{user.userFullName}</h1>
        <div className={styles.badges}>
          <RoleBadge role={user.role} isAr={isAr} />
          <UserStatusBadge status={user.status} isAr={isAr} />
        </div>
        <div className={styles.details}>
          {details.map(([icon, text]) => (
            <span key={icon}><i className={icon} />{text}</span>
          ))}
        </div>
      </div>

      <div className={styles.actions}>
        <Button variant="outline" icon="fa-solid fa-pen" onClick={onEdit}>{tr("Edit", "تعديل")}</Button>
        <Button variant="danger" icon="fa-solid fa-trash" onClick={onDelete}>{tr("Delete", "حذف")}</Button>
      </div>
    </Card>
  );
}
