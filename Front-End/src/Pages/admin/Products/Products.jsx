import { useLang } from "../../../hooks/useLang";
import { useFilters, matchesText } from "../../../hooks/useFilters";
import { usePagination } from "../../../hooks/usePagination";
import { Button, PageHeader, Pagination } from "../../../components/admin/ui";
import ProductsTable from "../../../components/admin/products/ProductsTable";
import ProductModals from "../../../components/admin/products/ProductModals";
import { useProductCrud } from "../../../components/admin/products/useProductCrud";
import ProductsStats from "./ProductsStats";
import ProductsFilters from "./ProductsFilters";

const PAGE_SIZE = 8;

const isOut = (p) => p.outOfStock || p.count === 0;

export default function Products() {
  const { tr } = useLang();
  const crud = useProductCrud();
  const [filters, setFilter] = useFilters({ search: "", category: "all", stock: "all" });

  const filtered = crud.products.filter(
    (p) =>
      matchesText(filters.search, p.name, p.nameAr, p.category) &&
      (filters.category === "all" || p.category === filters.category) &&
      (filters.stock === "all" || (filters.stock === "out") === isOut(p)),
  );
  const { page, setPage, totalPages, pageItems, from, to } = usePagination(filtered, PAGE_SIZE);

  return (
    <>
      <PageHeader
        icon="fa-solid fa-box"
        title={tr("Products", "المنتجات")}
        subtitle={tr("Manage and monitor all products in your store", "إدارة ومتابعة جميع منتجات متجرك")}
      >
        <Button icon="fa-solid fa-plus" onClick={crud.openAdd}>{tr("Add Product", "إضافة منتج")}</Button>
      </PageHeader>

      <ProductsStats products={crud.products} />
      <ProductsFilters filters={filters} setFilter={setFilter} categories={crud.categories} />
      <ProductsTable products={pageItems} onEdit={crud.openEdit} onDelete={crud.askDelete} />

      <Pagination
        page={page}
        totalPages={totalPages}
        onChange={setPage}
        summary={tr(`Showing ${from} to ${to} of ${filtered.length} products`, `عرض ${from} إلى ${to} من ${filtered.length} منتج`)}
      />

      <ProductModals crud={crud} />
    </>
  );
}
