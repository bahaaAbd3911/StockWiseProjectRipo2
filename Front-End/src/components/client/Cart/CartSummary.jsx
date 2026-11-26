import { useNavigate } from "react-router";

import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import { formatMoney } from "../../../utils/format";
import CouponBox from "./CouponBox";
import styles from "./Cart.module.css";

// Totals come from the backend (bundles, offers and the coupon are all counted there).
export default function CartSummary() {
  const navigate = useNavigate();
  const { tr } = useLang();
  const { cartInfo, cartCount } = useShop();

  return (
    <aside className={styles.summary}>
      <h3>{tr("Order Summary", "ملخص الطلب")}</h3>
      <div className={styles.line}>
        <span>{tr(`Items (${cartCount})`, `العناصر (${cartCount})`)}</span>
        <span>{formatMoney(cartInfo.subtotal)}</span>
      </div>
      {cartInfo.discount > 0 && (
        <div className={styles.line}>
          <span>{tr("Discount", "الخصم")}</span>
          <span className={styles.saving}>−{formatMoney(cartInfo.discount)}</span>
        </div>
      )}
      <div className={styles.line}>
        <span>{tr("Shipping", "الشحن")}</span>
        <span>{tr("Free", "مجاني")}</span>
      </div>
      <CouponBox />
      <div className={styles.total}>
        <span>{tr("Total", "الإجمالي")}</span>
        <span>{formatMoney(cartInfo.total)}</span>
      </div>
      <button type="button" className={styles.checkout} onClick={() => navigate("/client/PurchaseConfirmation")}>
        {tr("Proceed to Checkout", "متابعة إتمام الشراء")}
        <i className="fa-solid fa-arrow-right-long"></i>
      </button>
    </aside>
  );
}
