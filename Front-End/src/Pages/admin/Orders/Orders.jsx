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
import { exportOrdersCsv } from "./exportOrdersCsv";

const PAGE_SIZE = 8;
const DAY = 24 * 60 * 60 * 1000;

const matchesOrder = (order, f) =>
  matchesText(f.search, order.id, order.customerName, order.customerEmail, ...order.items.flatMap((i) => [i.name, i.nameAr])) &&
  (f.status === "all" || order.status === f.status) &&
  (f.payment === "all" || order.paymentMethod === f.payment) &&
  (f.date === "all" || Date.now() - new Date(order.createdAt).getTime() <= Number(f.date) * DAY);

export default function Orders() {
  const { tr, isAr } = useLang();
  const { orders, reload } = useAdminOrders();
  const { changeStatus } = useOrderStatusUpdate(reload);
  const [openOrderId, setOpenOrderId] = useState(null);
  const [filters, setFilter] = useFilters({ search: "", status: "all", payment: "all", date: "all" });

  const filtered = orders.filter((o) => matchesOrder(o, filters));
  const { page, setPage, totalPages, pageItems, from, to } = usePagination(filtered, PAGE_SIZE);

  return (
    <>
      <PageHeader
        icon="fa-solid fa-box"
        title={tr("Orders", "الطلبات")}
        subtitle={tr("Manage customer orders and track their status", "أدر طلبات العملاء وتابع حالتها")}
      >
        <Button variant="outline" icon="fa-solid fa-download" onClick={() => exportOrdersCsv(filtered)}>
          {tr("Export Orders", "تصدير الطلبات")}
        </Button>
      </PageHeader>

      <OrdersStats orders={orders} tr={tr} />
      <OrdersFilters filters={filters} setFilter={setFilter} isAr={isAr} tr={tr} />
      <OrdersTable orders={pageItems} onOpen={setOpenOrderId} />

      <Pagination
        page={page}
        totalPages={totalPages}
        onChange={setPage}
        summary={tr(`Showing ${from} to ${to} of ${filtered.length} orders`, `عرض ${from} إلى ${to} من ${filtered.length} طلب`)}
      />

      {openOrderId && (
        <OrderDetailsModal
          order={orders.find((o) => o.id === openOrderId)}
          onChangeStatus={changeStatus}
          onClose={() => setOpenOrderId(null)}
        />
      )}
    </>
  );
}
