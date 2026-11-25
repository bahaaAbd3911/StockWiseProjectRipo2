import { IconButton, StatusBadge, tableClasses } from "../ui";
import { finalPrice, formatMoney } from "../../../utils/format";
import styles from "./products.module.css";

function stockState(product, tr) {
  if (product.outOfStock || product.count === 0) return ["red", tr("Out of stock", "نفد المخزون")];
  if (product.count <= 5) return ["orange", tr("Low stock", "مخزون منخفض")];
  return ["green", tr("In stock", "متوفر")];
}

export default function ProductRow({ product, isAr, tr, onEdit, onDelete }) {
  const [tone, label] = stockState(product, tr);

  return (
    <tr className={tableClasses.clickableRow} onClick={() => onEdit(product)}>
      <td>
        <div className={styles.productCell}>
          <img src={product.image} alt="" />
          <div>
            <strong>{isAr ? product.nameAr : product.name}</strong>
            <small>#{product.id}</small>
          </div>
        </div>
      </td>
      <td>{isAr ? product.categoryAr : product.category}</td>
      <td>
        {formatMoney(finalPrice(product))}
        {product.discountValue > 0 && <small className={styles.discount}> -{product.discountValue}%</small>}
      </td>
      <td>{product.count}</td>
      <td>
        <i className="fa-solid fa-star" style={{ color: "#f59e0b" }} /> {product.rating}
      </td>
      <td>
        <StatusBadge tone={tone}>{label}</StatusBadge>
      </td>
      <td>
        <div className={tableClasses.rowActions}>
          <IconButton icon="fa-solid fa-pen" label={tr("Edit", "تعديل")} onClick={() => onEdit(product)} />
          <IconButton icon="fa-solid fa-trash" label={tr("Delete", "حذف")} danger onClick={() => onDelete(product)} />
        </div>
      </td>
    </tr>
  );
}
