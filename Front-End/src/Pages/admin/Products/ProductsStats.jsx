import { useLang } from "../../../hooks/useLang";
import { StatCard, StatsGrid } from "../../../components/admin/ui";

export default function ProductsStats({ products }) {
  const { tr } = useLang();
  const out = products.filter((p) => p.outOfStock || p.count === 0).length;
  const low = products.filter((p) => !p.outOfStock && p.count > 0 && p.count <= 5).length;

  return (
    <StatsGrid>
      <StatCard icon="fa-solid fa-box" label={tr("Total Products", "إجمالي المنتجات")} value={products.length} />
      <StatCard icon="fa-solid fa-circle-check" tone="blue" label={tr("In Stock", "متوفر")} value={products.length - out} />
      <StatCard icon="fa-solid fa-triangle-exclamation" tone="orange" label={tr("Low Stock", "مخزون منخفض")} value={low} />
      <StatCard icon="fa-solid fa-circle-xmark" tone="red" label={tr("Out of Stock", "نفد المخزون")} value={out} />
    </StatsGrid>
  );
}
