// Import Style
import { formContainerClass } from "../ShippingForm/ShippingForm.module.css";
import {
  paymentMethodsConatinerClass,
  btnContainerClass,
  NightBtnContainerClass,
  errorMsgClass,
} from "./PaymentMethod.module.css";
// Import Icons
import creditCard from "../../../../assets/images/Icon/contactless.png";

// Import Components
import FormHeader from "../FormHeader";
import Method from "./Method";

// Import Images
import visa from "../../../../assets/images/Icon/visa.png";
import paypal from "../../../../assets/images/Icon/paypal.png";
import dollar from "../../../../assets/images/Icon/dollar.png";

// Import Hooks
import { useContext, useEffect, useState } from "react";
import { AuthContext } from "../../../../context/AuthContext";
import { useShop } from "../../../../hooks/useShop";
import { fetchMyDebt } from "../../../../api/sales";
import { formatMoney } from "../../../../utils/format";

function PaymentMethod({ setSelectedStage}) {
  const { lang, orderStages, dark, orderFormFields} =
    useContext(AuthContext);

  const { cartInfo } = useShop();
  // Pay later is only offered to customers the store gave a credit limit.
  const [debt, setDebt] = useState(null);
  useEffect(() => {
    let current = true;
    fetchMyDebt()
      .then((res) => current && setDebt(res.account))
      .catch(() => {});
    return () => {
      current = false;
    };
  }, []);

  const methods = [
    ...(debt?.payLaterAllowed
      ? [
          {
            src: dollar,
            title: "Pay Later",
            arbTitle: "الدفع لاحقاً",
            subTitle: `Added to your balance (available credit ${formatMoney(debt.availableCredit)})`,
            arbSubTitle: `يُضاف إلى رصيدك (الرصيد المتاح ${formatMoney(debt.availableCredit)})`,
            isAvailable: debt.availableCredit >= cartInfo.total,
          },
        ]
      : []),
    {
      src: dollar,
      title: "Cash On Delivery",
      arbTitle: "الدفع عند الاستلام",
      subTitle: "Pay When You Receive Your Order",
      arbSubTitle: "ادفع عند استلام طلبك",
      isAvailable: true,
    },
    {
      src: visa,
      title: "Online Payment",
      arbTitle: "الدفع عبر الإنترنت",
      subTitle: "Pay Securely With Your Credit Or Debit Card",
      arbSubTitle:
        "ادفع بأمان باستخدام بطاقة الائتمان أو بطاقة الخصم الخاصة بك",
      isAvailable: false,
    },
    {
      src: paypal,
      title: "PayPal",
      arbTitle: "باي بال",
      subTitle: "Pay With Your PayPal Account",
      arbSubTitle: "ادفع باستخدام حساب PayPal الخاص بك",
      isAvailable: false,
    },
  ];
  return (
    <>
      <div className={formContainerClass}>
        <FormHeader
          icon={creditCard}
          title={"Payment Method"}
          arbTitle={`طريقة الدفع`}
          subTtile={"Choose Your Preferred Payment Method"}
          arbSubTtile={`اختر طريقة الدفع المفضلة لديك`}
        />

        <div className={paymentMethodsConatinerClass}>
          {methods.map((method, index) => (
            <Method method={method} key={index} />
          ))}
        </div>
        <span className={errorMsgClass}>
          {lang === "arb"
            ? orderFormFields[7].arbErrorMsg
            : orderFormFields[7].errorMsg}
        </span>
        <div
          className={`${btnContainerClass} ${!dark && NightBtnContainerClass}`}
        >
          <div
            onClick={() => {
              setSelectedStage(orderStages[0]);
            }}
          >
            <div>
              <i className="fa-solid fa-arrow-left"></i>
            </div>
            <div>
              {lang === "arb" ? "العودة إلى قسم الشحن" : "Back To Shipping"}
            </div>
          </div>
          <div
            onClick={() => {
              setSelectedStage(orderStages[2]);
            }}
          >
            <div>{lang === "arb" ? "تابع المراجعة" : "Continue To Review"}</div>
            <div>
              <i className="fa-solid fa-arrow-right"></i>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

export default PaymentMethod;
