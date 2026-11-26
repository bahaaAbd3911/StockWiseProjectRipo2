import { useContext } from "react";

import { formContainerClass } from "../ShippingForm/ShippingForm.module.css";
import { formContainersClass, btnContainerClass, NightBtnContainerClass } from "./PlaceOrder.module.css";
import listIcon from "../../../../assets/images/Icon/checklist.png";
import locationIcon from "../../../../assets/images/Icon/location.png";
import creditIcon from "../../../../assets/images/Icon/contactless.png";
import boxIcon from "../../../../assets/images/Icon/box.png";

import { AuthContext } from "../../../../context/AuthContext";
import { useShop } from "../../../../hooks/useShop";
import { useLang } from "../../../../hooks/useLang";
import { formatMoney } from "../../../../utils/format";
import FormHeader from "../FormHeader";
import ReviewCard from "./ReviewCard";
import { usePlaceOrder } from "../usePlaceOrder";

// Final step: review everything, then submit the order to the database.
function PlaceOrder({ setSelectedStage }) {
  const { dark, orderStages, orderFormFields: f } = useContext(AuthContext);
  const { cartInfo, cartCount } = useShop();
  const { tr } = useLang();
  const { placeOrder, placing } = usePlaceOrder();
  const total = cartInfo.total;
  const edit = tr("Edit", "تعديل");

  return (
    <div className={formContainerClass}>
      <FormHeader
        icon={listIcon}
        title={"Review Your Order"}
        arbTitle={"راجع طلبك"}
        subTtile={"Check the details below, then place your order."}
        arbSubTtile={"تحقق من التفاصيل أدناه ثم قدّم طلبك."}
      />

      <div className={formContainersClass}>
        <ReviewCard
          icon={locationIcon}
          title={tr("Shipping Information", "معلومات الشحن")}
          lines={[
            [tr("Name", "الاسم"), f[0].data],
            [tr("Phone", "الهاتف"), f[1].data],
            [tr("Email", "البريد"), f[2].data],
            [tr("Address", "العنوان"), [f[3].data, f[4].data, f[5].data].filter(Boolean).join(", ")],
            [tr("Delivery", "التوصيل"), f[6].data],
          ]}
          onEdit={() => setSelectedStage(orderStages[0])}
          editLabel={edit}
        />
        <ReviewCard
          icon={creditIcon}
          title={tr("Payment Method", "طريقة الدفع")}
          lines={[[tr("Method", "الطريقة"), f[7].data]]}
          onEdit={() => setSelectedStage(orderStages[1])}
          editLabel={edit}
        />
        <ReviewCard
          icon={boxIcon}
          title={tr("Order Items", "عناصر الطلب")}
          lines={[
            [tr("Items", "العناصر"), cartCount],
            ...(cartInfo.coupon ? [[tr("Coupon", "القسيمة"), cartInfo.coupon.code]] : []),
            ...(cartInfo.discount > 0 ? [[tr("Discount", "الخصم"), `−${formatMoney(cartInfo.discount)}`]] : []),
            [tr("Total", "الإجمالي"), formatMoney(total)],
          ]}
        />
      </div>

      <div className={`${btnContainerClass} ${!dark ? NightBtnContainerClass : ""}`}>
        <div onClick={() => setSelectedStage(orderStages[1])}>
          <div>
            <i className="fa-solid fa-arrow-left"></i>
          </div>
          <div>{tr("Back To Payment", "العودة إلى قسم الدفع")}</div>
        </div>

        <div onClick={placeOrder} aria-disabled={placing} style={placing ? { opacity: 0.6, pointerEvents: "none" } : undefined}>
          <div>
            {placing ? tr("Placing order…", "جارٍ تقديم الطلب…") : tr(`Place Order (${formatMoney(total)})`, `تأكيد الطلب (${formatMoney(total)})`)}
          </div>
          <div>
            <i className={`fa-solid ${placing ? "fa-spinner fa-spin" : "fa-lock"}`}></i>
          </div>
        </div>
      </div>
    </div>
  );
}

export default PlaceOrder;
