import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import { PAYMENT_METHODS } from "../../../constants/paymentMethods";
import { formatDate, formatMoney, pickLabel } from "../../../utils/format";
import styles from "./ClientOrdersPage.module.css";

// Expanded part of an order card: items, delivery info, and cancel (pending only).
export default function OrderDetails({ order }) {
  const { tr, isAr } = useLang();
  const { cancelOrder } = useShop();

  return (
    <div className={styles.orderDetailsClass}>
      {order.items.map((item) => (
        <div className={styles.orderItemRowClass} key={item.id}>
          <img src={item.image} alt="" />
          <div>
            <strong>{isAr ? item.nameAr : item.name}</strong>
            <small>
              {item.quantity} × {formatMoney(item.unitPrice)}
              {item.color && ` · ${item.color}`}
            </small>
          </div>
          <strong>{formatMoney(item.unitPrice * item.quantity)}</strong>
        </div>
      ))}

      <div className={styles.orderMetaClass}>
        <span>
          <i className="fa-solid fa-location-dot"></i> {order.recipientName} — {order.address}
        </span>
        <span>
          <i className="fa-solid fa-wallet"></i> {pickLabel(PAYMENT_METHODS[order.paymentMethod], isAr)}
        </span>
        <span>
          <i className="fa-solid fa-clock"></i> {formatDate(order.createdAt, isAr, true)}
        </span>
      </div>

      {order.status === "pending" && (
        <button type="button" className={styles.cancelOrderBtnClass} onClick={() => cancelOrder(order.id)}>
          <i className="fa-solid fa-ban"></i> {tr("Cancel Order", "إلغاء الطلب")}
        </button>
      )}
    </div>
  );
}
