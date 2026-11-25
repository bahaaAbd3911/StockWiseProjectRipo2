import { Avatar, IconButton, tableClasses } from "../ui";
import { formatDate, formatMoney, orderTotal } from "../../../utils/format";
import OrderStatusBadge from "./OrderStatusBadge";
import PaymentMethod from "./PaymentMethod";

export default function OrderRow({ order, isAr, tr, onOpen }) {
  const open = () => onOpen(order.id);

  return (
    <tr className={tableClasses.clickableRow} onClick={open}>
      <td><strong>#{order.id}</strong></td>
      <td>
        <span style={{ display: "flex", alignItems: "center", gap: 10 }}>
          <Avatar name={order.customerName} size={34} />
          {order.customerName}
        </span>
      </td>
      <td>{formatMoney(orderTotal(order))}</td>
      <td><PaymentMethod method={order.paymentMethod} isAr={isAr} /></td>
      <td><OrderStatusBadge status={order.status} isAr={isAr} /></td>
      <td>{formatDate(order.createdAt, isAr, true)}</td>
      <td>
        <IconButton icon="fa-solid fa-eye" label={tr("View details", "عرض التفاصيل")} onClick={open} />
      </td>
    </tr>
  );
}
