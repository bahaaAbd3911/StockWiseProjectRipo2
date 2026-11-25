// Import Style
import {
  paymentMethodConatinerClass,
  leftSecConatinerClass,
  rightSecConatinerClass,
  methodTitleClass,
  nightPaymentMethodConatinerClass,
  activeMethodClass,
} from "./PaymentMethod.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../../context/AuthContext";

// Import Component
import Icon from "../../../../components/client/Icon/Icon";

function Method({ method }) {
  const { lang, dark, orderFormFields, setorderFormFields } =
    useContext(AuthContext);
  return (
    <>
      <div
        className={`${paymentMethodConatinerClass} ${!dark && nightPaymentMethodConatinerClass} ${orderFormFields[7].data === method.title ? activeMethodClass : ""}`}
        aria-disabled={!method.isAvailable}
        style={method.isAvailable ? undefined : { opacity: 0.55, cursor: "not-allowed" }}
        onClick={() => {
          // Online payment needs a payment gateway, which isn't set up yet.
          if (!method.isAvailable) return;
          setorderFormFields((prev) => {
            let updated = [...prev];
            updated[7].data = method.title;
            updated[7].flag = false;
            updated[7].errorMsg = "";
            updated[7].arbErrorMsg = "";
            return updated;
          });
        }}
      >
        <div className={leftSecConatinerClass}>
          <Icon iconSrc={method.src} />
          <div className={methodTitleClass}>
            <div>{lang === "arb" ? method.arbTitle : method.title}</div>
            <div>{lang === "arb" ? method.arbSubTitle : method.subTitle}</div>
          </div>
        </div>
        <div className={rightSecConatinerClass}>
          {lang === "arb" && method.isAvailable && "متاح"}
          {lang !== "arb" && method.isAvailable && "Available"}
          {lang === "arb" && !method.isAvailable && "غير متاح"}
          {lang !== "arb" && !method.isAvailable && "Not Available Yet"}
        </div>
      </div>
    </>
  );
}

export default Method;
