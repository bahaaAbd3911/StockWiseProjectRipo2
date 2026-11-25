import { useLang } from "../../../hooks/useLang";
import { SearchInput, SelectInput, toolbarClass } from "../../../components/admin/ui";

export default function ProductsFilters({ filters, setFilter, categories }) {
  const { tr, isAr } = useLang();

  return (
    <div className={toolbarClass}>
      <SearchInput
        value={filters.search}
        onChange={(v) => setFilter("search", v)}
        placeholder={tr("Search products…", "ابحث عن منتج…")}
      />
      <SelectInput
        value={filters.category}
        onChange={(v) => setFilter("category", v)}
        options={[
          { value: "all", label: tr("All Categories", "كل التصنيفات") },
          ...categories.map((c) => ({ value: c.name, label: isAr ? c.nameAr : c.name })),
        ]}
      />
      <SelectInput
        value={filters.stock}
        onChange={(v) => setFilter("stock", v)}
        options={[
          { value: "all", label: tr("All Stock", "كل الحالات") },
          { value: "in", label: tr("In Stock", "متوفر") },
          { value: "out", label: tr("Out of Stock", "نفد المخزون") },
        ]}
      />
    </div>
  );
}
