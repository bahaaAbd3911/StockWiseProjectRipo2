import { Card, SearchInput, SelectInput } from "../../../../components/admin/ui";
import styles from "../Categories.module.css";

export default function CategorySidebar({ filters, setFilter, tr }) {
  return (
    <Card className={styles.sidebar}>
      <SearchInput
        value={filters.search}
        onChange={(v) => setFilter("search", v)}
        placeholder={tr("Search categories…", "ابحث عن تصنيف…")}
      />

      <h4>{tr("Filter by Status", "التصفية حسب الحالة")}</h4>
      <SelectInput
        value={filters.status}
        onChange={(v) => setFilter("status", v)}
        options={[
          { value: "all", label: tr("All Categories", "كل التصنيفات") },
          { value: "active", label: tr("Active", "نشط") },
          { value: "inactive", label: tr("Inactive", "غير نشط") },
        ]}
      />

      <h4>{tr("Sort by", "الترتيب حسب")}</h4>
      <SelectInput
        value={filters.sort}
        onChange={(v) => setFilter("sort", v)}
        options={[
          { value: "newest", label: tr("Newest First", "الأحدث أولاً") },
          { value: "oldest", label: tr("Oldest First", "الأقدم أولاً") },
          { value: "name", label: tr("Name (A–Z)", "الاسم (أ–ي)") },
          { value: "products", label: tr("Most Products", "الأكثر منتجات") },
        ]}
      />

      <div className={styles.promo}>
        <i className="fa-solid fa-folder-tree" />
        <h3>{tr("Organize Your Store", "نظّم متجرك")}</h3>
        <p>
          {tr(
            "Keep your categories well organized for a better shopping experience.",
            "حافظ على تنظيم التصنيفات لتجربة تسوق أفضل.",
          )}
        </p>
      </div>
    </Card>
  );
}
