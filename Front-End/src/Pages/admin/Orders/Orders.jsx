import { useState } from "react";

import { useLang } from "../../../hooks/useLang";
import { useFilters, matchesText } from "../../../hooks/useFilters";
import { usePagination } from "../../../hooks/usePagination";
import { useAdminOrders } from "../../../hooks/useAdminOrders";
import { useOrderStatusUpdate } from "../../../hooks/useOrderStatusUpdate";
import { Button, PageHeader, Pagination } from "../../../components/admin/ui";
import OrdersTable from "../../../components/admin/orders/OrdersTable";
import OrderDetailsModal from "../../../components/admin/orders/OrderDetailsModal";
import OrdersStats from "./components/OrdersStats";
import OrdersFilters from "./components/OrdersFilters";
import { downloadOrdersCsv } from "../../../api/admin";
import { useApiAction } from "../../../hooks/useApiAction";

const PAGE_SIZE = 8;
const DAY = 24 * 60 * 60 * 1000;

// YYYY-MM-DD of the first day a "last N days" filter covers (local date).
const sinceDay = (days) => {
  const d = new Date(Date.now() - Number(days) * DAY);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
};

const matchesOrder = (order, f) =>
  matchesText(
    f.search,
    order.id,
    order.customerName,
    order.customerEmail,
    ...order.items.flatMap((i) => [i.name, i.nameAr]),
  ) &&
  (f.status === "all" || order.status === f.status) &&
  (f.payment === "all" || order.paymentMethod === f.payment) &&
  (f.date === "all" ||
    Date.now() - new Date(order.createdAt).getTime() <= Number(f.date) * DAY);

export default function Orders() {
  const { tr, isAr } = useLang();
  const { orders, reload } = useAdminOrders();
  const { changeStatus } = useOrderStatusUpdate(reload);
  const [openOrderId, setOpenOrderId] = useState(null);
  const [filters, setFilter] = useFilters({
    search: "",
    status: "all",
    payment: "all",
    date: "all",
  });
  const { run, busy } = useApiAction();

  // The backend builds the file from every matching order (status, payment
  // and date filters; the free-text search is not applied).
  const exportCsv = () =>
    run(() =>
      downloadOrdersCsv({
        status: filters.status === "all" ? "" : filters.status,
        paymentMethod: filters.payment === "all" ? "" : filters.payment,
        from: filters.date === "all" ? "" : sinceDay(filters.date),
      }),
    );

  const filtered = orders.filter((o) => matchesOrder(o, filters));
  const { page, setPage, totalPages, pageItems, from, to } = usePagination(
    filtered,
    PAGE_SIZE,
  );

  return (
    <>
      <PageHeader
        icon="fa-solid fa-box"
        title={tr("Orders", "الطلبات")}
        subtitle={tr(
          "Manage customer orders and track their status",
          "أدر طلبات العملاء وتابع حالتها",
        )}
      >
        <Button
          variant="outline"
          icon="fa-solid fa-download"
          disabled={busy}
          onClick={exportCsv}
        >
          {tr("Export Orders", "تصدير الطلبات")}
        </Button>
      </PageHeader>

      <OrdersStats orders={orders} tr={tr} />
      <OrdersFilters
        filters={filters}
        setFilter={setFilter}
        isAr={isAr}
        tr={tr}
      />
      <OrdersTable orders={pageItems} onOpen={setOpenOrderId} />

      <Pagination
        page={page}
        totalPages={totalPages}
        onChange={setPage}
        summary={tr(
          `Showing ${from} to ${to} of ${filtered.length} orders`,
          `عرض ${from} إلى ${to} من ${filtered.length} طلب`,
        )}
      />

      {openOrderId && (
        <OrderDetailsModal
          order={orders.find((o) => o.id === openOrderId)}
          onChangeStatus={changeStatus}
          onNoteSaved={reload}
          onClose={() => setOpenOrderId(null)}
        />
      )}
    </>
  );
}
