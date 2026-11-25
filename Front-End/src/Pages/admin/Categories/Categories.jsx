import { useLang } from "../../../hooks/useLang";
import { useFilters, matchesText } from "../../../hooks/useFilters";
import { usePagination } from "../../../hooks/usePagination";
import { Button, PageHeader, Pagination } from "../../../components/admin/ui";
import { useCategoryCrud, isActiveCategory } from "./useCategoryCrud";
import CategoryStats from "./components/CategoryStats";
import CategoryGrid from "./components/CategoryGrid";
import CategorySidebar from "./components/CategorySidebar";
import CategoryModals from "./components/CategoryModals";
import styles from "./Categories.module.css";

const PAGE_SIZE = 8;

const SORTERS = {
  newest: (a, b) => b.id - a.id,
  oldest: (a, b) => a.id - b.id,
  name: (a, b) => a.name.localeCompare(b.name),
};

export default function Categories() {
  const { tr, isAr } = useLang();
  const crud = useCategoryCrud();
  const [filters, setFilter] = useFilters({ search: "", status: "all", sort: "newest" });

  const sorter = SORTERS[filters.sort] ?? ((a, b) => crud.countFor(b) - crud.countFor(a));
  const visible = crud.categories
    .filter(
      (c) =>
        matchesText(filters.search, c.name, c.nameAr) &&
        (filters.status === "all" || (filters.status === "active") === isActiveCategory(c)),
    )
    .sort(sorter);
  const { page, setPage, totalPages, pageItems, from, to } = usePagination(visible, PAGE_SIZE);

  return (
    <>
      <PageHeader
        icon="fa-solid fa-folder"
        title={tr("Categories Management", "إدارة التصنيفات")}
        subtitle={tr("Manage your store categories easily", "أدر تصنيفات متجرك بسهولة")}
      >
        <Button icon="fa-solid fa-plus" onClick={crud.openAdd}>{tr("Add Category", "إضافة تصنيف")}</Button>
      </PageHeader>

      <CategoryStats categories={crud.categories} tr={tr} />

      <div className={styles.layout}>
        <div>
          <CategoryGrid
            categories={pageItems}
            countFor={crud.countFor}
            isAr={isAr}
            tr={tr}
            onEdit={crud.openEdit}
            onDelete={crud.askDelete}
          />
          <Pagination
            page={page}
            totalPages={totalPages}
            onChange={setPage}
            summary={tr(`Showing ${from} to ${to} of ${visible.length} categories`, `عرض ${from} إلى ${to} من ${visible.length} تصنيف`)}
          />
        </div>
        <CategorySidebar filters={filters} setFilter={setFilter} tr={tr} />
      </div>

      <CategoryModals crud={crud} isAr={isAr} tr={tr} />
    </>
  );
}
