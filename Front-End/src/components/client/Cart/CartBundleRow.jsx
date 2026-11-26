import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import { formatMoney } from "../../../utils/format";
import styles from "./Cart.module.css";

// One bundle in the cart: a single price for several products.
export default function CartBundleRow({ bundle }) {
  const { tr, isAr } = useLang();
  const { updateBundleQuantity, removeBundle } = useShop();

  return (
    <div className={styles.row}>
      {bundle.image ? (
        <img src={bundle.image} alt="" />
      ) : (
        <div className={styles.bundleIcon}>
          <i className="fa-solid fa-boxes-stacked"></i>
        </div>
      )}

      <div className={styles.info}>
        <strong>
          <span className={styles.tag}>{tr("Bundle", "باقة")}</span> {isAr ? bundle.nameAr || bundle.name : bundle.name}
        </strong>
        <small>{bundle.contents}</small>
        <small>{formatMoney(bundle.unitPrice)}</small>
        {!bundle.isAvailable && <small className={styles.problem}>{bundle.unavailableReason}</small>}
      </div>

      <div className={styles.controls}>
        <div className={styles.stepper}>
          <button
            type="button"
            aria-label={tr("Decrease", "إنقاص")}
            disabled={bundle.quantity <= 1}
            onClick={() => updateBundleQuantity(bundle.id, bundle.quantity - 1)}
          >
            −
          </button>
          <span>{bundle.quantity}</span>
          <button
            type="button"
            aria-label={tr("Increase", "زيادة")}
            onClick={() => updateBundleQuantity(bundle.id, bundle.quantity + 1)}
          >
            +
          </button>
        </div>
        <span className={styles.lineTotal}>{formatMoney(bundle.lineTotal)}</span>
        <button type="button" className={styles.remove} onClick={() => removeBundle(bundle.id)}>
          <i className="fa-solid fa-trash"></i> {tr("Remove", "حذف")}
        </button>
      </div>
    </div>
  );
}
