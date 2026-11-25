// Pure helpers that turn API data into dashboard numbers.
import { orderTotal } from "../../../utils/format";

const DAY = 24 * 60 * 60 * 1000;
const isRevenue = (order) => order.status !== "cancelled";

// Compares the last 30 days with the 30 days before, e.g. { change: "12.5%", positive: true }.
function trend(items, dateOf, valueOf = () => 1) {
  const now = Date.now();
  let current = 0;
  let previous = 0;
  for (const item of items) {
    const age = now - new Date(dateOf(item)).getTime();
    if (age <= 30 * DAY) current += valueOf(item);
    else if (age <= 60 * DAY) previous += valueOf(item);
  }
  if (previous === 0) return { change: current ? "new" : "0%", positive: true };
  const pct = ((current - previous) / previous) * 100;
  return { change: `${Math.abs(pct).toFixed(1)}%`, positive: pct >= 0 };
}

export function buildStats({ users, orders, products }) {
  const revenueOrders = orders.filter(isRevenue);
  const revenue = revenueOrders.reduce((sum, o) => sum + orderTotal(o), 0);

  return [
    { key: "users", value: users.length.toLocaleString(), ...trend(users, (u) => u.joinedAt) },
    { key: "orders", value: orders.length.toLocaleString(), ...trend(orders, (o) => o.createdAt) },
    { key: "products", value: products.length.toLocaleString(), ...trend(products, (p) => p.createdAt) },
    {
      key: "revenue",
      value: `${revenue.toFixed(2)} JOD`,
      ...trend(revenueOrders, (o) => o.createdAt, orderTotal),
    },
  ];
}

// Best sellers by quantity sold in non-cancelled orders.
export function topProducts(orders, products, limit = 5) {
  const sold = new Map();
  for (const order of orders.filter(isRevenue)) {
    for (const item of order.items) {
      sold.set(item.productID, (sold.get(item.productID) ?? 0) + item.quantity);
    }
  }
  return [...sold]
    .map(([id, qty]) => ({ product: products.find((p) => p.id === id), sold: qty }))
    .filter((row) => row.product)
    .sort((a, b) => b.sold - a.sold)
    .slice(0, limit);
}

// Mixed feed of new orders and new accounts, newest first.
export function recentActivity(orders, users, limit = 6) {
  const events = [
    ...orders.map((o) => ({ name: o.customerName, action: "newOrder", date: o.createdAt })),
    ...users.map((u) => ({ name: u.userFullName, action: "newUser", date: u.joinedAt })),
  ];
  return events.sort((a, b) => new Date(b.date) - new Date(a.date)).slice(0, limit);
}

// Revenue per bucket for the sales chart: last 7 days, 8 weeks, or 6 months.
export function salesSeries(orders, period, isAr) {
  const now = new Date();
  const locale = isAr ? "ar-JO" : "en-US";
  let buckets;

  if (period === "daily") {
    buckets = Array.from({ length: 7 }, (_, i) => {
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate() - (6 - i));
      return { start, end: new Date(start.getTime() + DAY), label: start.toLocaleDateString(locale, { weekday: "short" }) };
    });
  } else if (period === "weekly") {
    buckets = Array.from({ length: 8 }, (_, i) => {
      const end = new Date(now.getTime() - (7 - i) * 7 * DAY + DAY);
      const start = new Date(end.getTime() - 7 * DAY);
      return { start, end, label: start.toLocaleDateString(locale, { month: "short", day: "numeric" }) };
    });
  } else {
    buckets = Array.from({ length: 6 }, (_, i) => {
      const start = new Date(now.getFullYear(), now.getMonth() - (5 - i), 1);
      return { start, end: new Date(start.getFullYear(), start.getMonth() + 1, 1), label: start.toLocaleDateString(locale, { month: "short" }) };
    });
  }

  return buckets.map((b) => ({
    label: b.label,
    value: orders
      .filter((o) => isRevenue(o) && new Date(o.createdAt) >= b.start && new Date(o.createdAt) < b.end)
      .reduce((sum, o) => sum + orderTotal(o), 0),
  }));
}

export function timeAgo(date, t) {
  const minutes = Math.floor((Date.now() - new Date(date).getTime()) / 60000);
  if (minutes < 1) return t.justNow;
  if (minutes < 60) return `${minutes} ${t.minutesAgo}`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} ${t.hoursAgo}`;
  return `${Math.floor(hours / 24)} ${t.daysAgo}`;
}
