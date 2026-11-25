import { useState } from "react";
import { paginate } from "../utils/format";

// Paginates a list; the page is clamped so shrinking the list never shows an empty page.
export function usePagination(list, pageSize) {
  const [requestedPage, setPage] = useState(1);
  const totalPages = Math.max(1, Math.ceil(list.length / pageSize));
  const page = Math.min(requestedPage, totalPages);

  return {
    page,
    setPage,
    totalPages,
    pageItems: paginate(list, page, pageSize),
    from: list.length ? (page - 1) * pageSize + 1 : 0,
    to: Math.min(page * pageSize, list.length),
  };
}
