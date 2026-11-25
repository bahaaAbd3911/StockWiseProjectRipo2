import { finalPrice, formatMoney } from "../../../utils/format";
import styles from "./Item.module.css";

// Shows the discounted price, and the original price struck through when discounted.
export default function PriceTag({ product }) {
  return (
    <div className={styles.itemPriceClass}>
      <span>{formatMoney(finalPrice(product))}</span>
      {product.discountValue > 0 && <span className={styles.oldPriceClass}>{formatMoney(product.price)}</span>}
    </div>
  );
}
