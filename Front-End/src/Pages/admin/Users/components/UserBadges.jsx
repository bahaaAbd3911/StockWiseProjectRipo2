import { StatusBadge } from "../../../../components/admin/ui";
import { USER_ROLES, USER_STATUSES } from "../../../../constants/userRoles";
import { pickLabel } from "../../../../utils/format";

export function RoleBadge({ role, isAr }) {
  const info = USER_ROLES[role] ?? USER_ROLES.customer;
  return (
    <StatusBadge tone={info.tone} icon="fa-solid fa-shield-halved">
      {pickLabel(info, isAr)}
    </StatusBadge>
  );
}

export function UserStatusBadge({ status, isAr }) {
  const info = USER_STATUSES[status] ?? USER_STATUSES.active;
  return <StatusBadge tone={info.tone}>{pickLabel(info, isAr)}</StatusBadge>;
}
