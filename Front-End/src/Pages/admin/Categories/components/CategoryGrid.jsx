import { EmptyState } from "../../../../components/admin/ui";
import CategoryCard from "./CategoryCard";
import styles from "../Categories.module.css";

export default function CategoryGrid({ categories, countFor, isAr, tr, onEdit, onDelete }) {
  if (categories.length === 0) {
    return <EmptyState icon="fa-solid fa-folder-open" text={tr("No categories found", "لا توجد تصنيفات")} />;
  }

  return (
    <div className={styles.grid}>
      {categories.map((category) => (
        <CategoryCard
          key={category.id}
          category={category}
          count={countFor(category)}
          isAr={isAr}
          tr={tr}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
}
