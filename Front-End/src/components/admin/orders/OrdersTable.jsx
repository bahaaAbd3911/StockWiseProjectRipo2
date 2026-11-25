import { useLang } from "../../../hooks/useLang";
import { Card, EmptyState, tableClasses } from "../ui";
import OrderRow from "./OrderRow";

export default function OrdersTable({ orders, onOpen }) {
  const { tr, isAr } = useLang();

  if (orders.length === 0) {
    return (
      <Card>
        <EmptyState icon="fa-solid fa-receipt" text={tr("No orders found", "لا توجد طلبات")} />
      </Card>
    );
  }

  const headers = [
    tr("Order #", "رقم الطلب"),
    tr("Customer", "العميل"),
    tr("Total Amount", "المبلغ الإجمالي"),
    tr("Payment Method", "طريقة الدفع"),
    tr("Status", "الحالة"),
    tr("Order Date", "تاريخ الطلب"),
    tr("Actions", "الإجراءات"),
  ];

  return (
    <Card className={tableClasses.wrap}>
      <table className={tableClasses.table}>
        <thead>
          <tr>{headers.map((h) => <th key={h}>{h}</th>)}</tr>
        </thead>
        <tbody>
          {orders.map((order) => (
            <OrderRow key={order.id} order={order} isAr={isAr} tr={tr} onOpen={onOpen} />
          ))}
        </tbody>
      </table>
    </Card>
  );
}
