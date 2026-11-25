import { useState } from "react";

// Small helper for a page's filter object: setFilter("search", "abc").
export function useFilters(initial) {
  const [filters, setFilters] = useState(initial);
  const setFilter = (key, value) => setFilters((prev) => ({ ...prev, [key]: value }));
  return [filters, setFilter];
}

export const matchesText = (query, ...fields) => {
  const q = query.trim().toLowerCase();
  return !q || fields.some((field) => String(field ?? "").toLowerCase().includes(q));
};
