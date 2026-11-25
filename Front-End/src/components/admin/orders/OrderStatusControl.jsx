import { useState } from "react";

import { Button, selectClass } from "../ui";
import { ORDER_STATUS_LABELS, nextStatuses } from "../../../constants/orderStatus";
import { pickLabel } from "../../../utils/format";
import OrderStatusBadge from "./OrderStatusBadge";
import styles from "./orders.module.css";

// Lets the admin move an order to one of its allowed next statuses.
export default function OrderStatusControl({ order, isAr, tr, onChange }) {
  const options = nextStatuses(order.status);
  const [selected, setSelected] = useState(options[0] ?? "");

  return (
    <section className={styles.section}>
      <h3>{tr("Order Status", "حالة الطلب")}</h3>

      <div className={styles.statusControl}>
        <OrderStatusBadge status={order.status} isAr={isAr} />

        {options.length === 0 ? (
          <p className={styles.finalNote}>{tr("This order is closed and can't change.", "هذا الطلب مغلق ولا يمكن تغييره.")}</p>
        ) : (
          <>
            <i className={`fa-solid ${isAr ? "fa-arrow-left" : "fa-arrow-right"}`} />
            <select className={selectClass} value={selected} onChange={(e) => setSelected(e.target.value)}>
              {options.map((s) => (
                <option key={s} value={s}>{pickLabel(ORDER_STATUS_LABELS[s], isAr)}</option>
              ))}
            </select>
            <Button icon="fa-solid fa-check" onClick={() => onChange(selected)}>
              {tr("Update Status", "تحديث الحالة")}
            </Button>
          </>
        )}
      </div>
    </section>
  );
}
