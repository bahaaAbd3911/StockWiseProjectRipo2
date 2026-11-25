import { StatCard, StatsGrid } from "../../../../components/admin/ui";
import { OPEN_STATUSES } from "../../../../constants/orderStatus";

export default function OrdersStats({ orders, tr }) {
  const count = (...statuses) => orders.filter((o) => statuses.includes(o.status)).length;

  return (
    <StatsGrid>
      <StatCard icon="fa-solid fa-cart-shopping" label={tr("Total Orders", "إجمالي الطلبات")} value={orders.length} />
      <StatCard icon="fa-solid fa-square-check" tone="blue" label={tr("Delivered", "تم التوصيل")} value={count("delivered")} />
      <StatCard icon="fa-solid fa-clock" tone="orange" label={tr("In Progress", "قيد التنفيذ")} value={count(...OPEN_STATUSES)} />
      <StatCard icon="fa-solid fa-xmark" tone="red" label={tr("Cancelled", "ملغاة")} value={count("cancelled")} />
    </StatsGrid>
  );
}
