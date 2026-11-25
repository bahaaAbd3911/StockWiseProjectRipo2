import { useLang } from "../../../hooks/useLang";
import { orderedProductCountClass, sectionLabelClass } from "./ProductDetails.module.css";

// Quantity picker limited to 1..stock.
function OrderedProduct({ productData, quantity, setQuantity }) {
  const { tr } = useLang();
  const max = Math.max(productData.count, 1);

  return (
    <div>
      <p className={sectionLabelClass}>{tr("Quantity", "الكمية")}</p>
      <div className={orderedProductCountClass}>
        <button
          type="button"
          aria-label={tr("Decrease", "إنقاص")}
          disabled={quantity <= 1}
          onClick={() => setQuantity(quantity - 1)}
        >
          −
        </button>
        <span aria-live="polite">{quantity}</span>
        <button
          type="button"
          aria-label={tr("Increase", "زيادة")}
          disabled={quantity >= max}
          onClick={() => setQuantity(quantity + 1)}
        >
          +
        </button>
      </div>
    </div>
  );
}

export default OrderedProduct;
