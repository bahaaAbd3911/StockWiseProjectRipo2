import { FaEye } from "react-icons/fa";
import { useNavigate } from "react-router";

import { formatDate, formatMoney, orderTotal } from "../../../../utils/format";
import Status from "./Status";
import styles from "../Dashboard.module.css";

export default function OrdersPanel({ t, orders, isAr }) {
  const navigate = useNavigate();

  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3>{t.recentOrders}</h3>
        <button className={styles.viewAll} onClick={() => navigate("/admin/orders")}>
          {t.viewAll} <FaEye />
        </button>
      </div>

      {orders.length === 0 ? (
        <p className={styles.panelSubtitle}>{t.noData}</p>
      ) : (
        <div className={styles.tableWrap}>
          <table className={styles.dataTable}>
            <thead>
              <tr>
                <th>{t.order}</th>
                <th>{t.customer}</th>
                <th>{t.date}</th>
                <th>{t.amount}</th>
                <th>{t.status}</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id}>
                  <td className={styles.orderId}>#{order.id}</td>
                  <td>{order.customerName}</td>
                  <td>{formatDate(order.createdAt, isAr)}</td>
                  <td>{formatMoney(orderTotal(order))}</td>
                  <td>
                    <Status status={order.status} t={t} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
