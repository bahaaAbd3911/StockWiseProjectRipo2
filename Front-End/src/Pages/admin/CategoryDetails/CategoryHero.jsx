import { Button, Card, StatusBadge } from "../../../components/admin/ui";
import styles from "./CategoryDetails.module.css";

export default function CategoryHero({ category, count, isAr, tr, onAdd }) {
  const active = category.isActive !== false;

  return (
    <Card className={styles.hero}>
      <span className={styles.heroIcon}>
        <i className={category.icon} />
      </span>

      <div className={styles.heroText}>
        <h1>{isAr ? category.nameAr : category.name}</h1>
        <p>{tr(`${count} products in this category`, `${count} منتج في هذا التصنيف`)}</p>
        <StatusBadge tone={active ? "green" : "blue"}>
          {active ? tr("Active", "نشط") : tr("Inactive", "غير نشط")}
        </StatusBadge>
      </div>

      <Button icon="fa-solid fa-plus" onClick={onAdd}>{tr("Add Product", "إضافة منتج")}</Button>
    </Card>
  );
}
