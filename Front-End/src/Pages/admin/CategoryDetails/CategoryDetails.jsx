import { useState } from "react";
import { Link, Navigate, useParams } from "react-router";

import { useLang } from "../../../hooks/useLang";
import { matchesText } from "../../../hooks/useFilters";
import { useCatalog } from "../../../hooks/useCatalog";
import { EmptyState, SearchInput, toolbarClass } from "../../../components/admin/ui";
import ProductsTable from "../../../components/admin/products/ProductsTable";
import ProductModals from "../../../components/admin/products/ProductModals";
import { useProductCrud } from "../../../components/admin/products/useProductCrud";
import CategoryHero from "./CategoryHero";
import styles from "./CategoryDetails.module.css";

export default function CategoryDetails() {
  const { categoryId } = useParams();
  const { tr, isAr } = useLang();
  const crud = useProductCrud();
  const [search, setSearch] = useState("");

  const { status } = useCatalog();
  const category = crud.categories.find((c) => String(c.id) === categoryId);

  // Wait for the catalogue before deciding the category doesn't exist (direct page loads).
  if (!category && status === "loading") {
    return <EmptyState icon="fa-solid fa-spinner fa-spin" text={tr("Loading…", "جارٍ التحميل…")} />;
  }
  if (!category) return <Navigate to="/admin/categories" replace />;

  const inCategory = crud.products.filter((p) => p.category === category.name);
  const visible = inCategory.filter((p) => matchesText(search, p.name, p.nameAr));

  return (
    <>
      <Link to="/admin/categories" className={styles.back}>
        <i className={`fa-solid ${isAr ? "fa-arrow-right" : "fa-arrow-left"}`} />
        {tr("Back to categories", "العودة إلى التصنيفات")}
      </Link>

      <CategoryHero category={category} count={inCategory.length} isAr={isAr} tr={tr} onAdd={crud.openAdd} />

      <div className={toolbarClass}>
        <SearchInput value={search} onChange={setSearch} placeholder={tr("Search in this category…", "ابحث في هذا التصنيف…")} />
      </div>

      <ProductsTable products={visible} onEdit={crud.openEdit} onDelete={crud.askDelete} />

      <ProductModals crud={crud} lockedCategory={category.name} />
    </>
  );
}
