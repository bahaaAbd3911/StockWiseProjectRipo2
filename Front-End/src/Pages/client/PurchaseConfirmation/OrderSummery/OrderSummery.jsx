import {
  OrderSummeryContainerClass,
  OrderSummeryHeadContainerClass,
  costsContainerClass,
  costClass,
  TotalClass,
  productsContainerClass,
  productClass,
  productImageClass,
  productInfoClass,
  productPriceAndCountClass,
  counterClass,
} from "./OrderSummery.module.css";

import { useContext, useEffect } from "react";

import { AuthContext } from "../../../../context/AuthContext";
import { useShop } from "../../../../hooks/useShop";
import { useLang } from "../../../../hooks/useLang";
import { formatMoney } from "../../../../utils/format";
import ShippingLine from "../../../../components/client/Cart/ShippingLine";

// Right-hand summary on the checkout page, built from the real cart.
function OrderSummery() {
  const {
    cart,
    cartInfo,
    cartCount,
    updateCartQuantity,
    removeFromCart,
    updateBundleQuantity,
    removeBundle,
    setShippingCity,
  } = useShop();
  const { tr, isAr } = useLang();
  const { orderFormFields } = useContext(AuthContext);
  const city = orderFormFields[3].data; // checkout field 3 = city

  // Shipping depends on the delivery city: ask the backend again when it changes.
  useEffect(() => {
    setShippingCity(city).catch(() => {});
  }, [city, setShippingCity]);

  return (
    <div className={OrderSummeryContainerClass}>
      <div className={OrderSummeryHeadContainerClass}>
        <div>
          <div>
            <i className="fa-solid fa-cart-shopping"></i>
          </div>
          <div>{tr("Order Summary", "ملخص الطلب")}</div>
        </div>
        <div>{tr(`${cartCount} Items`, `${cartCount} عناصر`)}</div>
      </div>

      <div className={productsContainerClass}>
        {cart.map((item) => (
          <div className={productClass} key={item.id}>
            <div>
              <div className={productImageClass}>
                <img src={item.image} alt="" style={{ objectFit: "cover" }} />
              </div>
              <div className={productInfoClass}>
                <div>{isAr ? item.nameAr : item.name}</div>
                {item.color && <small>{item.color}</small>}
                <div>{formatMoney(item.unitPrice)}</div>
              </div>
            </div>
            <div className={productPriceAndCountClass}>
              <div className={counterClass}>
                <div
                  role="button"
                  aria-label={tr("Decrease", "إنقاص")}
                  onClick={() =>
                    item.quantity > 1
                      ? updateCartQuantity(item.id, item.quantity - 1)
                      : removeFromCart(item.id)
                  }
                >
                  −
                </div>
                <div>{item.quantity}</div>
                <div
                  role="button"
                  aria-label={tr("Increase", "زيادة")}
                  onClick={() =>
                    item.quantity < item.stock &&
                    updateCartQuantity(item.id, item.quantity + 1)
                  }
                >
                  +
                </div>
              </div>
            </div>
          </div>
        ))}
        {cartInfo.bundles.map((b) => (
          <div className={productClass} key={`bundle-${b.id}`}>
            <div>
              <div className={productImageClass}>
                {b.image ? (
                  <img src={b.image} alt="" style={{ objectFit: "cover" }} />
                ) : (
                  <i className="fa-solid fa-boxes-stacked"></i>
                )}
              </div>
              <div className={productInfoClass}>
                <div>{isAr ? b.nameAr || b.name : b.name}</div>
                <small>{b.contents}</small>
                <div>{formatMoney(b.unitPrice)}</div>
              </div>
            </div>
            <div className={productPriceAndCountClass}>
              <div className={counterClass}>
                <div
                  role="button"
                  aria-label={tr("Decrease", "إنقاص")}
                  onClick={() =>
                    b.quantity > 1
                      ? updateBundleQuantity(b.id, b.quantity - 1)
                      : removeBundle(b.id)
                  }
                >
                  −
                </div>
                <div>{b.quantity}</div>
                <div
                  role="button"
                  aria-label={tr("Increase", "زيادة")}
                  onClick={() => updateBundleQuantity(b.id, b.quantity + 1)}
                >
                  +
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className={costsContainerClass}>
        <div className={costClass}>
          <div>{tr("Subtotal", "المجموع الفرعي")}</div>
          <div>{formatMoney(cartInfo.subtotal)}</div>
        </div>
        {cartInfo.discount > 0 && (
          <div className={costClass}>
            <div>
              {tr("Discount", "الخصم")}
              {cartInfo.coupon ? ` (${cartInfo.coupon.code})` : ""}
            </div>
            <div>−{formatMoney(cartInfo.discount)}</div>
          </div>
        )}
        <ShippingLine
          cartInfo={cartInfo}
          rowClass={costClass}
          cell="div"
          label={tr("Shipping Fee", "رسوم الشحن")}
        />
      </div>
      <div className={TotalClass}>
        <div>{tr("Total", "الإجمالي")}</div>
        <div>{formatMoney(cartInfo.total)}</div>
      </div>
    </div>
  );
}

export default OrderSummery;
