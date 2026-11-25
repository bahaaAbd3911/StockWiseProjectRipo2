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

import { useShop } from "../../../../hooks/useShop";
import { useLang } from "../../../../hooks/useLang";
import { cartSubtotal, formatMoney } from "../../../../utils/format";

// Right-hand summary on the checkout page, built from the real cart.
function OrderSummery() {
  const { cart, cartCount, updateCartQuantity, removeFromCart } = useShop();
  const { tr, isAr } = useLang();
  const subtotal = cartSubtotal(cart);

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
                  onClick={() => (item.quantity > 1 ? updateCartQuantity(item.id, item.quantity - 1) : removeFromCart(item.id))}
                >
                  −
                </div>
                <div>{item.quantity}</div>
                <div
                  role="button"
                  aria-label={tr("Increase", "زيادة")}
                  onClick={() => item.quantity < item.stock && updateCartQuantity(item.id, item.quantity + 1)}
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
          <div>{formatMoney(subtotal)}</div>
        </div>
        <div className={costClass}>
          <div>{tr("Shipping Fee", "رسوم الشحن")}</div>
          <div>{tr("Free", "مجاني")}</div>
        </div>
      </div>
      <div className={TotalClass}>
        <div>{tr("Total", "الإجمالي")}</div>
        <div>{formatMoney(subtotal)}</div>
      </div>
    </div>
  );
}

export default OrderSummery;
