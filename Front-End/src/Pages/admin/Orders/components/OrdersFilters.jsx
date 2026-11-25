import { SearchInput, SelectInput, toolbarClass } from "../../../../components/admin/ui";
import { ORDER_STATUSES, ORDER_STATUS_LABELS } from "../../../../constants/orderStatus";
import { PAYMENT_METHODS } from "../../../../constants/paymentMethods";
import { pickLabel } from "../../../../utils/format";

export default function OrdersFilters({ filters, setFilter, isAr, tr }) {
  return (
    <div className={toolbarClass}>
      <SearchInput
        value={filters.search}
        onChange={(v) => setFilter("search", v)}
        placeholder={tr("Search by order ID, customer name or product…", "ابحث برقم الطلب أو اسم العميل أو المنتج…")}
      />
      <SelectInput
        value={filters.status}
        onChange={(v) => setFilter("status", v)}
        options={[
          { value: "all", label: tr("All Statuses", "كل الحالات") },
          ...ORDER_STATUSES.map((s) => ({ value: s, label: pickLabel(ORDER_STATUS_LABELS[s], isAr) })),
        ]}
      />
      <SelectInput
        value={filters.payment}
        onChange={(v) => setFilter("payment", v)}
        options={[
          { value: "all", label: tr("All Payment Methods", "كل طرق الدفع") },
          ...Object.entries(PAYMENT_METHODS).map(([value, info]) => ({ value, label: pickLabel(info, isAr) })),
        ]}
      />
      <SelectInput
        value={filters.date}
        onChange={(v) => setFilter("date", v)}
        options={[
          { value: "all", label: tr("Any Date", "أي تاريخ") },
          { value: "7", label: tr("Last 7 days", "آخر 7 أيام") },
          { value: "30", label: tr("Last 30 days", "آخر 30 يومًا") },
          { value: "90", label: tr("Last 90 days", "آخر 90 يومًا") },
        ]}
      />
    </div>
  );
}
