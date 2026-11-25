import { Card, EmptyState, tableClasses } from "../../../../components/admin/ui";
import UserRow from "./UserRow";

export default function UsersTable({ users, startIndex, isAr, tr, onEdit, onDelete }) {
  if (users.length === 0) {
    return (
      <Card>
        <EmptyState icon="fa-solid fa-user-slash" text={tr("No users found", "لا يوجد مستخدمون")} />
      </Card>
    );
  }

  const headers = ["#", tr("Name", "الاسم"), tr("Email", "البريد"), tr("Phone", "الهاتف"), tr("Role", "الدور"), tr("Status", "الحالة"), tr("Joined At", "تاريخ الانضمام"), tr("Actions", "الإجراءات")];

  return (
    <Card className={tableClasses.wrap}>
      <table className={tableClasses.table}>
        <thead>
          <tr>{headers.map((h) => <th key={h}>{h}</th>)}</tr>
        </thead>
        <tbody>
          {users.map((user, i) => (
            <UserRow
              key={user.userID}
              user={user}
              index={startIndex + i}
              isAr={isAr}
              tr={tr}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          ))}
        </tbody>
      </table>
    </Card>
  );
}
