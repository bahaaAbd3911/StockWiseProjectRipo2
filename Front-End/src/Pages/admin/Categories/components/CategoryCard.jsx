import { useNavigate } from "react-router";

import { Card, IconButton, StatusBadge } from "../../../../components/admin/ui";
import { isActiveCategory } from "../useCategoryCrud";
import styles from "../Categories.module.css";

export default function CategoryCard({ category, count, isAr, tr, onEdit, onDelete }) {
  const navigate = useNavigate();
  const active = isActiveCategory(category);

  return (
    <Card className={styles.card} onClick={() => navigate(`/admin/categories/${category.id}`)}>
      <span className={styles.cardIcon}>
        <i className={category.icon} />
      </span>
      <h3>{isAr ? category.nameAr : category.name}</h3>
      <p className={styles.count}>{tr(`${count} Products`, `${count} منتج`)}</p>

      <StatusBadge tone={active ? "green" : "blue"}>
        {active ? tr("Active", "نشط") : tr("Inactive", "غير نشط")}
      </StatusBadge>

      <div className={styles.cardActions}>
        <IconButton icon="fa-solid fa-pen" label={tr("Edit", "تعديل")} onClick={() => onEdit(category)} />
        <IconButton icon="fa-solid fa-trash" label={tr("Delete", "حذف")} danger onClick={() => onDelete(category)} />
      </div>
    </Card>
  );
}
