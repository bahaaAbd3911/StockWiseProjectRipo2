import { StatCard, StatsGrid } from "../../../../components/admin/ui";
import { isActiveCategory } from "../useCategoryCrud";

export default function CategoryStats({ categories, tr }) {
  const total = categories.length;
  const active = categories.filter(isActiveCategory).length;
  const percent = (n) => `${total ? Math.round((n / total) * 100) : 0}%`;

  return (
    <StatsGrid>
      <StatCard icon="fa-solid fa-folder" label={tr("Total Categories", "إجمالي التصنيفات")} value={total} />
      <StatCard
        icon="fa-solid fa-eye"
        label={tr("Active Categories", "التصنيفات النشطة")}
        value={active}
        trend={percent(active)}
      />
      <StatCard
        icon="fa-solid fa-pause"
        tone="gray"
        label={tr("Inactive Categories", "التصنيفات غير النشطة")}
        value={total - active}
        trend={percent(total - active)}
        down
      />
    </StatsGrid>
  );
}
