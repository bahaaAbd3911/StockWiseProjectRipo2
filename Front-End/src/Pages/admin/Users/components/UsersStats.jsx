import { StatCard, StatsGrid } from "../../../../components/admin/ui";

export default function UsersStats({ users, tr }) {
  const count = (role) => users.filter((u) => u.role === role).length;

  return (
    <StatsGrid>
      <StatCard icon="fa-solid fa-users" label={tr("Total Users", "إجمالي المستخدمين")} value={users.length} />
      <StatCard icon="fa-solid fa-user" label={tr("Customers", "العملاء")} value={count("customer")} />
      <StatCard icon="fa-solid fa-user-slash" label={tr("Inactive", "غير النشطين")} value={users.filter((u) => u.status === "inactive").length} />
      <StatCard icon="fa-solid fa-user-gear" tone="purple" label={tr("Admins", "المدراء")} value={count("admin")} />
    </StatsGrid>
  );
}
