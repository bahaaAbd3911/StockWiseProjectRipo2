import { finalPrice, formatMoney } from "../../../utils/format";
import styles from "./Item.module.css";

// Shows the price the customer pays (own discount or live offer), and the
// original price struck through when it is lower.
export default function PriceTag({ product }) {
  const price = finalPrice(product);
  return (
    <div className={styles.itemPriceClass}>
      <span>{formatMoney(price)}</span>
      {price < product.price && <span className={styles.oldPriceClass}>{formatMoney(product.price)}</span>}
    </div>
  );
}
