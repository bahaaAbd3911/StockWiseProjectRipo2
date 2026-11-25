import { useNavigate } from "react-router";

import { Avatar, IconButton, tableClasses } from "../../../../components/admin/ui";
import { formatDate } from "../../../../utils/format";
import { RoleBadge, UserStatusBadge } from "./UserBadges";

export default function UserRow({ user, index, isAr, tr, onEdit, onDelete }) {
  const navigate = useNavigate();
  const openProfile = () => navigate(`/admin/users/${user.userID}`);

  return (
    <tr className={tableClasses.clickableRow} onClick={openProfile}>
      <td>{index}</td>
      <td>
        <span style={{ display: "flex", alignItems: "center", gap: 10 }}>
          <Avatar name={user.userFullName} />
          {user.userFullName}
        </span>
      </td>
      <td>{user.userEmail}</td>
      <td>{user.phone || "—"}</td>
      <td><RoleBadge role={user.role} isAr={isAr} /></td>
      <td><UserStatusBadge status={user.status} isAr={isAr} /></td>
      <td>{formatDate(user.joinedAt, isAr)}</td>
      <td>
        <div className={tableClasses.rowActions}>
          <IconButton icon="fa-solid fa-eye" label={tr("View profile", "عرض الملف")} onClick={openProfile} />
          <IconButton icon="fa-solid fa-pen" label={tr("Edit", "تعديل")} onClick={() => onEdit(user)} />
          <IconButton icon="fa-solid fa-trash" label={tr("Delete", "حذف")} danger onClick={() => onDelete(user)} />
        </div>
      </td>
    </tr>
  );
}
