import { useLang } from "../../../hooks/useLang";
import { Button, Modal } from "../ui";
import OrderInfoGrid from "./OrderInfoGrid";
import OrderItemsList from "./OrderItemsList";
import OrderStatusControl from "./OrderStatusControl";
import OrderNoteCard from "./OrderNoteCard";
import styles from "./orders.module.css";

// `order` comes from the parent's (API-loaded) list, so it updates after a status change.
// onNoteSaved: reload the list so the saved note shows.
export default function OrderDetailsModal({
  order,
  onChangeStatus,
  onNoteSaved,
  onClose,
}) {
  const { tr, isAr } = useLang();
  if (!order) return null;

  return (
    <Modal
      wide
      title={tr(`Order #${order.id}`, `الطلب #${order.id}`)}
      onClose={onClose}
    >
      <div className={styles.modalActions}>
        <Button
          variant="outline"
          icon="fa-solid fa-print"
          onClick={() =>
            window.open(`/invoice/${order.id}?view=admin`, "_blank", "noopener")
          }
        >
          {tr("Print invoice / packing slip", "طباعة الفاتورة / قائمة التغليف")}
        </Button>
      </div>
      <OrderStatusControl
        key={order.status}
        order={order}
        isAr={isAr}
        tr={tr}
        onChange={(status) => onChangeStatus(order.id, status)}
      />
      <OrderInfoGrid order={order} isAr={isAr} tr={tr} />
      <OrderItemsList order={order} isAr={isAr} tr={tr} />
      <OrderNoteCard
        key={order.id}
        order={order}
        isAr={isAr}
        tr={tr}
        onSaved={onNoteSaved}
      />
    </Modal>
  );
}
