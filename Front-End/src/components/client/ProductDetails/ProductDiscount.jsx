import { useLang } from "../../../hooks/useLang";
import { finalPrice, formatMoney } from "../../../utils/format";
import { productDiscountClass, newPriceClass, oldPriceClass, discountValueClass } from "./ProductDetails.module.css";

function ProductDiscount({ productData }) {
  const { tr, isAr } = useLang();
  const price = finalPrice(productData);
  const percent = productData.price > 0 ? Math.round((1 - price / productData.price) * 100) : 0;
  const offer = productData.offer; // the live offer that set the price, if any

  return (
    <div className={productDiscountClass}>
      <div>
        <span className={newPriceClass}>{formatMoney(price)}</span>
        {price < productData.price && (
          <>
            <span className={oldPriceClass}>{formatMoney(productData.price)}</span>
            <span className={discountValueClass}>
              {tr("Save", "وفّر")} {percent}%
              {offer && ` · ${isAr ? offer.name_ar || offer.name : offer.name}`}
            </span>
          </>
        )}
      </div>
    </div>
  );
}

export default ProductDiscount;
