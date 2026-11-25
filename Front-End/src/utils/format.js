export const formatMoney = (value) => `${Number(value || 0).toFixed(2)} JOD`;

export const formatDate = (iso, isAr, withTime = false) =>
  new Date(iso).toLocaleString(isAr ? "ar-JO" : "en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    ...(withTime && { hour: "2-digit", minute: "2-digit" }),
  });

export const finalPrice = (product) =>
  product.price * (1 - (Number(product.discountValue) || 0) / 100);

export const nextId = (list, key = "id") =>
  Math.max(0, ...list.map((item) => Number(item[key]) || 0)) + 1;

export const paginate = (list, page, size) =>
  list.slice((page - 1) * size, page * size);

export const initials = (name = "") =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0].toUpperCase())
    .join("");

// Prefer the total stored by the server; fall back to summing the items.
export const orderTotal = (order) =>
  order.total ?? order.items.reduce((sum, item) => sum + item.unitPrice * item.quantity, 0);

// Picks the English or Arabic text of a { en, ar } label.
export const pickLabel = (label, isAr) =>
  label ? (isAr ? label.ar : label.en) : "";

// Sum of a cart's line totals (cart items carry their discounted unitPrice).
export const cartSubtotal = (cart) => cart.reduce((sum, i) => sum + i.unitPrice * i.quantity, 0);
