import { useLang } from "../../../hooks/useLang";
import { Modal } from "../ui";
import OrderInfoGrid from "./OrderInfoGrid";
import OrderItemsList from "./OrderItemsList";
import OrderStatusControl from "./OrderStatusControl";

// `order` comes from the parent's (API-loaded) list, so it updates after a status change.
export default function OrderDetailsModal({ order, onChangeStatus, onClose }) {
  const { tr, isAr } = useLang();
  if (!order) return null;

  return (
    <Modal wide title={tr(`Order #${order.id}`, `الطلب #${order.id}`)} onClose={onClose}>
      <OrderStatusControl
        key={order.status}
        order={order}
        isAr={isAr}
        tr={tr}
        onChange={(status) => onChangeStatus(order.id, status)}
      />
      <OrderInfoGrid order={order} isAr={isAr} tr={tr} />
      <OrderItemsList order={order} isAr={isAr} tr={tr} />
    </Modal>
  );
}
