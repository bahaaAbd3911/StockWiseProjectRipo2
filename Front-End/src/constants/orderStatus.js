// Must match the statuses and transitions enforced by the backend (order_repository.go).
export const ORDER_STATUSES = ["pending", "confirmed", "processing", "shipped", "delivered", "cancelled"];

export const ORDER_STATUS_LABELS = {
  pending: { en: "Pending", ar: "قيد الانتظار" },
  confirmed: { en: "Confirmed", ar: "مؤكد" },
  processing: { en: "Processing", ar: "قيد التجهيز" },
  shipped: { en: "Shipped", ar: "تم الشحن" },
  delivered: { en: "Delivered", ar: "تم التوصيل" },
  cancelled: { en: "Cancelled", ar: "ملغي" },
};

export const ORDER_STATUS_TONES = {
  pending: "orange",
  confirmed: "blue",
  processing: "blue",
  shipped: "purple",
  delivered: "green",
  cancelled: "red",
};

export const ORDER_STATUS_ICONS = {
  pending: "fa-solid fa-clock",
  confirmed: "fa-solid fa-circle-check",
  processing: "fa-solid fa-gears",
  shipped: "fa-solid fa-truck",
  delivered: "fa-solid fa-bag-shopping",
  cancelled: "fa-solid fa-circle-xmark",
};

// Orders only move forward; delivered and cancelled orders are final.
const TRANSITIONS = {
  pending: ["confirmed", "cancelled"],
  confirmed: ["processing", "cancelled"],
  processing: ["shipped"],
  shipped: ["delivered"],
  delivered: [],
  cancelled: [],
};

export const nextStatuses = (status) => TRANSITIONS[status] ?? [];

// Orders that are still on their way to the customer.
export const OPEN_STATUSES = ["pending", "confirmed", "processing", "shipped"];
