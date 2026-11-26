import { useState } from "react";

import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import styles from "./Cart.module.css";

// Enter a coupon code, or see / remove the one already applied.
// The backend checks the code and works out the discount.
export default function CouponBox() {
  const { tr } = useLang();
  const { cartInfo, applyCoupon, removeCoupon } = useShop();
  const [code, setCode] = useState("");
  const [busy, setBusy] = useState(false);
  const coupon = cartInfo.coupon;

  const submit = async (e) => {
    e.preventDefault();
    if (!code.trim()) return;
    setBusy(true);
    if (await applyCoupon(code)) setCode("");
    setBusy(false);
  };

  if (coupon) {
    return (
      <div className={styles.coupon}>
        <div className={styles.line}>
          <span>
            <i className="fa-solid fa-ticket"></i> {coupon.code}
          </span>
          <button type="button" className={styles.remove} onClick={removeCoupon}>
            {tr("Remove", "إزالة")}
          </button>
        </div>
        {coupon.problem && <small className={styles.problem}>{coupon.problem}</small>}
      </div>
    );
  }

  return (
    <form className={styles.couponForm} onSubmit={submit}>
      <input
        value={code}
        onChange={(e) => setCode(e.target.value)}
        placeholder={tr("Coupon code", "رمز القسيمة")}
        maxLength={50}
      />
      <button type="submit" disabled={busy || !code.trim()}>
        {tr("Apply", "تطبيق")}
      </button>
    </form>
  );
}
