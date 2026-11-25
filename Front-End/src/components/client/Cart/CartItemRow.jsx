import { useNavigate } from "react-router";

import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import { formatMoney } from "../../../utils/format";
import styles from "./Cart.module.css";

export default function CartItemRow({ item }) {
  const navigate = useNavigate();
  const { tr, isAr } = useLang();
  const { updateCartQuantity, removeFromCart } = useShop();
  const open = () => navigate(`/client/ProductDetails/${item.productID}`);

  return (
    <div className={styles.row}>
      <img src={item.image} alt="" onClick={open} />

      <div className={styles.info}>
        <strong onClick={open}>{isAr ? item.nameAr : item.name}</strong>
        {item.color && (
          <small>
            {tr("Color", "اللون")}: {item.color}
          </small>
        )}
        <small>{formatMoney(item.unitPrice)}</small>
      </div>

      <div className={styles.controls}>
        <div className={styles.stepper}>
          <button
            type="button"
            aria-label={tr("Decrease", "إنقاص")}
            disabled={item.quantity <= 1}
            onClick={() => updateCartQuantity(item.id, item.quantity - 1)}
          >
            −
          </button>
          <span>{item.quantity}</span>
          <button
            type="button"
            aria-label={tr("Increase", "زيادة")}
            disabled={item.quantity >= item.stock}
            onClick={() => updateCartQuantity(item.id, item.quantity + 1)}
          >
            +
          </button>
        </div>
        <span className={styles.lineTotal}>{formatMoney(item.unitPrice * item.quantity)}</span>
        <button type="button" className={styles.remove} onClick={() => removeFromCart(item.id)}>
          <i className="fa-solid fa-trash"></i> {tr("Remove", "حذف")}
        </button>
      </div>
    </div>
  );
}
