import { formatDate } from "../../../utils/format";
import PaymentMethod from "./PaymentMethod";
import styles from "./orders.module.css";

export default function OrderInfoGrid({ order, isAr, tr }) {
  const info = [
    [tr("Customer", "العميل"), order.customerName],
    [tr("Email", "البريد"), order.customerEmail || "—"],
    [tr("Phone", "الهاتف"), order.customerPhone || "—"],
    [tr("Address", "العنوان"), order.address],
    [tr("Payment Method", "طريقة الدفع"), <PaymentMethod key="pm" method={order.paymentMethod} isAr={isAr} />],
    [tr("Order Date", "تاريخ الطلب"), formatDate(order.createdAt, isAr, true)],
  ];

  return (
    <section className={styles.section}>
      <h3>{tr("Order Information", "معلومات الطلب")}</h3>
      <div className={styles.infoGrid}>
        {info.map(([label, value]) => (
          <div key={label} className={styles.infoItem}>
            <span>{label}</span>
            <strong>{value}</strong>
          </div>
        ))}
      </div>
    </section>
  );
}
