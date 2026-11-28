import { useContext, useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router";

import { AuthContext } from "../../context/AuthContext";
import { fetchAdminOrder } from "../../api/admin";
import { fetchMyOrder } from "../../api/shop";
import { useBrand } from "../../hooks/useBrand";
import { useLang } from "../../hooks/useLang";
import { PAYMENT_METHODS } from "../../constants/paymentMethods";
import { ORDER_STATUS_LABELS } from "../../constants/orderStatus";
import { formatDate, formatMoney, pickLabel } from "../../utils/format";
import styles from "./Invoice.module.css";

// /invoice/:orderId - a printable invoice for one order.
//   customers: their own order (GET /api/orders/{id})
//   admins (?view=admin): any order of the store, plus the private note, so
//   the same page works as a packing slip.
// Access is decided by the backend; this page only shows what it returns.
export default function Invoice() {
  const { orderId } = useParams();
  const [params] = useSearchParams();
  const { ready, isSign, isAdmin } = useContext(AuthContext);
  const { tr, isAr } = useLang();
  const brand = useBrand();
  const asAdmin = params.get("view") === "admin" && isAdmin;
  const [state, setState] = useState({ order: null, error: "" });

  useEffect(() => {
    if (!ready || !isSign) return;
    let current = true;
    (asAdmin ? fetchAdminOrder(orderId) : fetchMyOrder(orderId)).then(
      (order) => current && setState({ order, error: "" }),
      (err) => current && setState({ order: null, error: err.message }),
    );
    return () => {
      current = false;
    };
  }, [ready, isSign, asAdmin, orderId]);

  if (ready && !isSign) {
    return (
      <div className={styles.page}>
        <p>
          {tr(
            "Please sign in to see this invoice.",
            "يرجى تسجيل الدخول لعرض هذه الفاتورة.",
          )}
        </p>
        <Link to="/signIn">{tr("Sign in", "تسجيل الدخول")}</Link>
      </div>
    );
  }
  const { order, error } = state;
  if (error) {
    return (
      <div className={styles.page}>
        <p className={styles.error}>{error}</p>
      </div>
    );
  }
  if (!order)
    return <div className={styles.page}>{tr("Loading…", "جارٍ التحميل…")}</div>;

  const name = (i) => (isAr ? i.nameAr || i.name : i.name);

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <button type="button" onClick={() => window.print()}>
          <i className="fa-solid fa-print" /> {tr("Print", "طباعة")}
        </button>
        <Link to={asAdmin ? "/admin/orders" : "/client/profile"}>
          {tr("Back", "رجوع")}
        </Link>
      </div>

      <article className={styles.invoice}>
        <header className={styles.head}>
          <div className={styles.brand}>
            {brand.logo && <img src={brand.logo} alt="" />}
            <strong>{brand.name}</strong>
          </div>
          <div className={styles.titleBlock}>
            <h1>{tr("Invoice", "فاتورة")}</h1>
            <p>
              {tr("Order", "الطلب")} #{order.id}
            </p>
            <p>{formatDate(order.createdAt, isAr, true)}</p>
          </div>
        </header>

        <section className={styles.parties}>
          <div>
            <h2>{tr("Customer", "العميل")}</h2>
            <p>{order.customerName}</p>
            {order.customerEmail && <p>{order.customerEmail}</p>}
          </div>
          <div>
            <h2>{tr("Deliver to", "التوصيل إلى")}</h2>
            <p>{order.recipientName}</p>
            <p>{order.customerPhone}</p>
            <p>{order.address}</p>
          </div>
          <div>
            <h2>{tr("Payment", "الدفع")}</h2>
            <p>
              {pickLabel(PAYMENT_METHODS[order.paymentMethod], isAr) ||
                order.paymentMethod}
            </p>
            <p>
              {tr("Status", "الحالة")}:{" "}
              {pickLabel(ORDER_STATUS_LABELS[order.status], isAr) ||
                order.status}
            </p>
          </div>
        </section>

        <table className={styles.items}>
          <thead>
            <tr>
              <th>{tr("Item", "المنتج")}</th>
              <th>{tr("Qty", "الكمية")}</th>
              <th>{tr("Unit price", "سعر الوحدة")}</th>
              <th>{tr("Total", "المجموع")}</th>
            </tr>
          </thead>
          <tbody>
            {order.items.map((i) => (
              <tr key={i.id}>
                <td>
                  {name(i)}
                  {i.color && <small> · {i.color}</small>}
                </td>
                <td>{i.quantity}</td>
                <td>{formatMoney(i.unitPrice)}</td>
                <td>{formatMoney(i.unitPrice * i.quantity)}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className={styles.totals}>
          <div>
            <span>{tr("Subtotal", "المجموع الفرعي")}</span>
            <span>{formatMoney(order.subtotal)}</span>
          </div>
          {order.discount > 0 && (
            <div>
              <span>
                {tr("Coupon", "القسيمة")}
                {order.couponCode ? ` ${order.couponCode}` : ""}
              </span>
              <span>−{formatMoney(order.discount)}</span>
            </div>
          )}
          <div>
            <span>{tr("Shipping", "الشحن")}</span>
            <span>
              {order.shipping > 0
                ? formatMoney(order.shipping)
                : tr("Free", "مجاني")}
            </span>
          </div>
          <div className={styles.grand}>
            <span>{tr("Total", "الإجمالي")}</span>
            <span>{formatMoney(order.total)}</span>
          </div>
        </div>

        {order.notes && (
          <section className={styles.note}>
            <h2>{tr("Delivery instructions", "تعليمات التوصيل")}</h2>
            <p>{order.notes}</p>
          </section>
        )}
        {asAdmin && order.adminNote && (
          <section className={`${styles.note} ${styles.private}`}>
            <h2>
              <i className="fa-solid fa-lock" />{" "}
              {tr(
                "Private note (not shown to the customer)",
                "ملاحظة خاصة (لا يراها العميل)",
              )}
            </h2>
            <p>{order.adminNote}</p>
          </section>
        )}

        <footer className={styles.foot}>
          {tr(
            `Thank you for shopping with ${brand.name}.`,
            `شكراً لتسوقك من ${brand.name}.`,
          )}
        </footer>
      </article>
    </div>
  );
}
