import { useLang } from "../../../hooks/useLang";
import { useFilters, matchesText } from "../../../hooks/useFilters";
import { usePagination } from "../../../hooks/usePagination";
import { Button, PageHeader, Pagination } from "../../../components/admin/ui";
import { useUserCrud } from "./useUserCrud";
import UsersStats from "./components/UsersStats";
import UsersFilters from "./components/UsersFilters";
import UsersTable from "./components/UsersTable";
import UserModals from "./components/UserModals";

const PAGE_SIZE = 8;
const DAY = 24 * 60 * 60 * 1000;

const joinedWithin = (user, days) =>
  days === "all" || Date.now() - new Date(user.joinedAt).getTime() <= Number(days) * DAY;

export default function Users() {
  const { tr, isAr } = useLang();
  const crud = useUserCrud();
  const [filters, setFilter] = useFilters({ search: "", role: "all", status: "all", joined: "all" });

  const filtered = crud.users.filter(
    (u) =>
      matchesText(filters.search, u.userFullName, u.userEmail, u.phone) &&
      (filters.role === "all" || u.role === filters.role) &&
      (filters.status === "all" || u.status === filters.status) &&
      joinedWithin(u, filters.joined),
  );
  const { page, setPage, totalPages, pageItems, from, to } = usePagination(filtered, PAGE_SIZE);

  return (
    <>
      <PageHeader
        icon="fa-solid fa-users"
        title={tr("Users Management", "إدارة المستخدمين")}
        subtitle={tr("Manage your store users and their roles", "أدر مستخدمي متجرك وأدوارهم")}
      >
        <Button icon="fa-solid fa-plus" onClick={crud.openAdd}>{tr("Add User", "إضافة مستخدم")}</Button>
      </PageHeader>

      <UsersStats users={crud.users} tr={tr} />
      <UsersFilters filters={filters} setFilter={setFilter} isAr={isAr} tr={tr} />
      <UsersTable users={pageItems} startIndex={from} isAr={isAr} tr={tr} onEdit={crud.openEdit} onDelete={crud.askDelete} />

      <Pagination
        page={page}
        totalPages={totalPages}
        onChange={setPage}
        summary={tr(`Showing ${from} to ${to} of ${filtered.length} users`, `عرض ${from} إلى ${to} من ${filtered.length} مستخدم`)}
      />

      <UserModals crud={crud} isAr={isAr} tr={tr} />
    </>
  );
}
