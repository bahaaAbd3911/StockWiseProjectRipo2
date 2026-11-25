import { orderTotal } from "../../../utils/format";

const escape = (value) => `"${String(value ?? "").replace(/"/g, '""')}"`;

// Downloads the given orders as a CSV file.
export function exportOrdersCsv(orders) {
  const header = ["Order", "Customer", "Email", "Items", "Total (JOD)", "Payment", "Status", "Date"];
  const rows = orders.map((o) => [
    o.id,
    o.customerName,
    o.customerEmail,
    o.items.map((i) => `${i.name} x${i.quantity}`).join("; "),
    orderTotal(o).toFixed(2),
    o.paymentMethod,
    o.status,
    o.createdAt,
  ]);

  const csv = [header, ...rows].map((row) => row.map(escape).join(",")).join("\n");
  // The leading BOM makes Excel read the Arabic text as UTF-8.
  const BOM = String.fromCharCode(0xfeff);
  const url = URL.createObjectURL(new Blob([BOM + csv], { type: "text/csv;charset=utf-8" }));
  const link = Object.assign(document.createElement("a"), { href: url, download: "orders.csv" });
  link.click();
  URL.revokeObjectURL(url);
}
