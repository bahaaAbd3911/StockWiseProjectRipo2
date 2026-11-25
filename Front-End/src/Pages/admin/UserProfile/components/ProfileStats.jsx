import { StatCard, StatsGrid } from "../../../../components/admin/ui";
import { formatMoney, orderTotal } from "../../../../utils/format";
import { OPEN_STATUSES } from "../../../../constants/orderStatus";

export default function ProfileStats({ orders, favoritesCount, tr }) {
  const spent = orders.filter((o) => o.status !== "cancelled").reduce((sum, o) => sum + orderTotal(o), 0);
  const open = orders.filter((o) => OPEN_STATUSES.includes(o.status)).length;

  return (
    <StatsGrid>
      <StatCard icon="fa-solid fa-bag-shopping" label={tr("Orders", "الطلبات")} value={orders.length} />
      <StatCard icon="fa-solid fa-sack-dollar" tone="blue" label={tr("Total Spent", "إجمالي الإنفاق")} value={formatMoney(spent)} />
      <StatCard icon="fa-solid fa-clock" tone="orange" label={tr("Open Orders", "طلبات مفتوحة")} value={open} />
      <StatCard icon="fa-solid fa-heart" tone="red" label={tr("Favorites", "المفضلة")} value={favoritesCount} />
    </StatsGrid>
  );
}
