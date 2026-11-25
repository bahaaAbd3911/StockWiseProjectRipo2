import { formatMoney, orderTotal } from "../../../utils/format";
import styles from "./orders.module.css";

export default function OrderItemsList({ order, isAr, tr }) {
  return (
    <section className={styles.section}>
      <h3>{tr(`Products (${order.items.length})`, `المنتجات (${order.items.length})`)}</h3>

      {order.items.map((item, i) => (
        <div key={`${item.productID}-${i}`} className={styles.item}>
          {item.image ? <img src={item.image} alt="" /> : <span />}
          <div className={styles.itemInfo}>
            <p>{isAr ? item.nameAr : item.name}</p>
            <small>
              {tr("Qty", "الكمية")}: {item.quantity}
              {item.color && ` · ${tr("Color", "اللون")}: ${item.color}`}
              {` · ${formatMoney(item.unitPrice)}`}
            </small>
          </div>
          <strong>{formatMoney(item.unitPrice * item.quantity)}</strong>
        </div>
      ))}

      <div className={styles.total}>
        <span>{tr("Total", "المجموع")}</span>
        <span>{formatMoney(orderTotal(order))}</span>
      </div>
    </section>
  );
}
