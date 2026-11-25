import styles from "./ui.module.css";

// Shows first/last pages, the current page and its neighbours, with ellipses between.
function pageNumbers(page, totalPages) {
  const wanted = new Set([1, totalPages, page - 1, page, page + 1]);
  const pages = [...wanted].filter((p) => p >= 1 && p <= totalPages).sort((a, b) => a - b);

  return pages.flatMap((p, i) => (i > 0 && p - pages[i - 1] > 1 ? ["…", p] : [p]));
}

export default function Pagination({ page, totalPages, onChange, summary }) {
  return (
    <div className={styles.pagination}>
      <span>{summary}</span>

      <div className={styles.pages}>
        <button className={styles.pageBtn} disabled={page <= 1} onClick={() => onChange(page - 1)} aria-label="Previous page">
          <i className="fa-solid fa-chevron-left" />
        </button>

        {pageNumbers(page, totalPages).map((p, i) =>
          p === "…" ? (
            <span key={`gap-${i}`} className={styles.pageBtn} style={{ border: "none" }}>…</span>
          ) : (
            <button
              key={p}
              className={`${styles.pageBtn} ${p === page ? styles.pageActive : ""}`}
              onClick={() => onChange(p)}
            >
              {p}
            </button>
          ),
        )}

        <button className={styles.pageBtn} disabled={page >= totalPages} onClick={() => onChange(page + 1)} aria-label="Next page">
          <i className="fa-solid fa-chevron-right" />
        </button>
      </div>
    </div>
  );
}
