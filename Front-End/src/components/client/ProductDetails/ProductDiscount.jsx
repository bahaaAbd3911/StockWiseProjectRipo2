import { useLang } from "../../../hooks/useLang";
import { finalPrice, formatMoney } from "../../../utils/format";
import { productDiscountClass, newPriceClass, oldPriceClass, discountValueClass } from "./ProductDetails.module.css";

function ProductDiscount({ productData }) {
  const { tr } = useLang();
  const hasDiscount = productData.discountValue > 0;

  return (
    <div className={productDiscountClass}>
      <div>
        <span className={newPriceClass}>{formatMoney(finalPrice(productData))}</span>
        {hasDiscount && (
          <>
            <span className={oldPriceClass}>{formatMoney(productData.price)}</span>
            <span className={discountValueClass}>
              {tr("Save", "وفّر")} {productData.discountValue}%
            </span>
          </>
        )}
      </div>
    </div>
  );
}

export default ProductDiscount;
