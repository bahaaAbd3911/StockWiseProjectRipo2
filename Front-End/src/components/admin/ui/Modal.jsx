import { useEffect } from "react";
import { createPortal } from "react-dom";

import { useLang } from "../../../hooks/useLang";
import IconButton from "./IconButton";
import "../../../layouts/AdminLayout/adminTheme.css";
import styles from "./ui.module.css";

export default function Modal({ title, onClose, footer, wide, children }) {
  const { dir, dark, tr } = useLang();

  useEffect(() => {
    const onKey = (e) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  // Portaled to <body>, so it re-declares the admin theme class itself.
  return createPortal(
    <div
      className={`adminTheme ${dark ? "adminThemeDark" : ""} ${styles.overlay}`}
      dir={dir}
      onMouseDown={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label={title}
        className={`${styles.modal} ${wide ? styles.modalWide : ""}`}
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className={styles.modalHeader}>
          <h2>{title}</h2>
          <IconButton icon="fa-solid fa-xmark" label={tr("Close", "إغلاق")} onClick={onClose} />
        </div>

        <div className={styles.modalBody}>{children}</div>

        {footer && <div className={styles.modalFooter}>{footer}</div>}
      </div>
    </div>,
    document.body,
  );
}
