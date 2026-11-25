import { SearchInput, SelectInput, toolbarClass } from "../../../../components/admin/ui";
import { USER_ROLES, USER_STATUSES } from "../../../../constants/userRoles";
import { pickLabel } from "../../../../utils/format";

export default function UsersFilters({ filters, setFilter, isAr, tr }) {
  const options = (map, allLabel) => [
    { value: "all", label: allLabel },
    ...Object.entries(map).map(([value, label]) => ({ value, label: pickLabel(label, isAr) })),
  ];

  return (
    <div className={toolbarClass}>
      <SearchInput
        value={filters.search}
        onChange={(v) => setFilter("search", v)}
        placeholder={tr("Search by name, email or phone…", "ابحث بالاسم أو البريد أو الهاتف…")}
      />
      <SelectInput value={filters.role} onChange={(v) => setFilter("role", v)} options={options(USER_ROLES, tr("All Roles", "كل الأدوار"))} />
      <SelectInput value={filters.status} onChange={(v) => setFilter("status", v)} options={options(USER_STATUSES, tr("All Statuses", "كل الحالات"))} />
      <SelectInput
        value={filters.joined}
        onChange={(v) => setFilter("joined", v)}
        options={[
          { value: "all", label: tr("Joined: All Time", "الانضمام: كل الأوقات") },
          { value: "7", label: tr("Last 7 days", "آخر 7 أيام") },
          { value: "30", label: tr("Last 30 days", "آخر 30 يومًا") },
          { value: "90", label: tr("Last 90 days", "آخر 90 يومًا") },
        ]}
      />
    </div>
  );
}
