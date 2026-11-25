import styles from "./ClientOrdersPage.module.css";

// The customer-facing order filters group the backend statuses.
export const ORDER_GROUPS = [
  { key: "all", name: "All Orders", arbName: "كل الطلبات", icon: "fa-solid fa-list", statuses: null },
  { key: "pending", name: "Pending", arbName: "قيد الانتظار", icon: "fa-solid fa-clock", statuses: ["pending", "confirmed"] },
  { key: "shipping", name: "On The Way", arbName: "في الطريق", icon: "fa-solid fa-truck", statuses: ["processing", "shipped"] },
  { key: "delivered", name: "Delivered", arbName: "تم التوصيل", icon: "fa-solid fa-circle-check", statuses: ["delivered"] },
  { key: "cancelled", name: "Cancelled", arbName: "ملغاة", icon: "fa-solid fa-ban", statuses: ["cancelled"] },
];

export const inGroup = (order, group) => !group.statuses || group.statuses.includes(order.status);

// Card color theme per status.
export const STATUS_STYLE = {
  pending: styles.PendingOrderContainerClass,
  confirmed: styles.PendingOrderContainerClass,
  processing: styles.ShippingOrderContainerClass,
  shipped: styles.ShippingOrderContainerClass,
  delivered: styles.DeliveredOrderContainerClass,
  cancelled: styles.CancelledOrderContainerClass,
};
