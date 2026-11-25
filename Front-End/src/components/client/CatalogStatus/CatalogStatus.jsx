import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import styles from "./CatalogStatus.module.css";

// Loading / error placeholder while products come from the API. Renders children when ready.
export default function CatalogStatus({ children }) {
  const { status, error, reload } = useCatalog();
  const { tr } = useLang();

  if (status === "ready") return children;

  if (status === "loading") {
    return (
      <div className={styles.box} role="status">
        <i className="fa-solid fa-spinner fa-spin"></i>
        <p>{tr("Loading products…", "جارٍ تحميل المنتجات…")}</p>
      </div>
    );
  }

  return (
    <div className={styles.box} role="alert">
      <i className="fa-solid fa-plug-circle-xmark"></i>
      <p>{tr("We couldn't load the store right now.", "تعذر تحميل المتجر الآن.")}</p>
      <small>{error}</small>
      <button onClick={reload}>{tr("Try again", "أعد المحاولة")}</button>
    </div>
  );
}
