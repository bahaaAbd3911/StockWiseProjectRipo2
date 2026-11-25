import { useNavigate } from "react-router";

import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import { cartSubtotal, formatMoney } from "../../../utils/format";
import styles from "./Cart.module.css";

export default function CartSummary() {
  const navigate = useNavigate();
  const { tr } = useLang();
  const { cart, cartCount } = useShop();
  const subtotal = cartSubtotal(cart);

  return (
    <aside className={styles.summary}>
      <h3>{tr("Order Summary", "ملخص الطلب")}</h3>
      <div className={styles.line}>
        <span>{tr(`Items (${cartCount})`, `العناصر (${cartCount})`)}</span>
        <span>{formatMoney(subtotal)}</span>
      </div>
      <div className={styles.line}>
        <span>{tr("Shipping", "الشحن")}</span>
        <span>{tr("Free", "مجاني")}</span>
      </div>
      <div className={styles.total}>
        <span>{tr("Total", "الإجمالي")}</span>
        <span>{formatMoney(subtotal)}</span>
      </div>
      <button type="button" className={styles.checkout} onClick={() => navigate("/client/PurchaseConfirmation")}>
        {tr("Proceed to Checkout", "متابعة إتمام الشراء")}
        <i className="fa-solid fa-arrow-right-long"></i>
      </button>
    </aside>
  );
}
