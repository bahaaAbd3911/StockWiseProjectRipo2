import { useContext, useEffect, useState } from "react";

import {
  PurchaseConfirmationContianerClass,
  arbPurchaseConfirmationContianerClass,
  FormContainerClass,
  rightSectionClass,
  leftSectionClass,
  gridContainerClass,
} from "./PurchaseConfirmation.module.css";
import { AuthContext } from "../../../context/AuthContext";
import { useShop } from "../../../hooks/useShop";

import UserSettings from "../../../components/common/UserSettings/UserSettings";
import SearchBar from "../../../components/common/SearchBar/SearchBar";
import SignInRequired from "../../../components/client/SignInRequired/SignInRequired";
import EmptyMessage from "../../../components/client/EmptyMessage/EmptyMessage";
import PurchaseConfirmationHeader from "./PurchaseConfirmationHeader";
import OrderData from "./ConfirmationStages/OrderData";
import ShippingForm from "./ShippingForm/ShippingForm";
import PaymentMethod from "./PaymentMethod/PaymentMethod";
import PlaceOrder from "./PlaceOrder/PlaceOrder";
import OrderSummery from "./OrderSummery/OrderSummery";
import cartImg from "../../../assets/images/ShoppingCart/cartLogo.png";

const STEPS = { Shipping: ShippingForm, Payment: PaymentMethod, Review: PlaceOrder };

// Fills name / phone / email from the signed-in user the first time checkout opens.
function usePrefillContact() {
  const { user, setorderFormFields } = useContext(AuthContext);
  useEffect(() => {
    if (!user) return;
    const values = [user.userFullName, user.phone, user.userEmail];
    setorderFormFields((prev) => prev.map((f, i) => (i < 3 && !f.data ? { ...f, data: values[i] ?? "" } : f)));
  }, [user, setorderFormFields]);
}

function PurchaseConfirmation() {
  const { lang, orderStages, isSign } = useContext(AuthContext);
  const { cart } = useShop();
  const [selectedStage, setSelectedStage] = useState(orderStages[0]);
  usePrefillContact();

  const Step = STEPS[selectedStage.label];

  let body;
  if (!isSign) body = <SignInRequired text={null} />;
  else if (cart.length === 0)
    body = (
      <EmptyMessage
        img={cartImg}
        title={"Your cart is empty"}
        arbTitle={"سلة التسوق فارغة"}
        subTitle={"Add some products to your cart before checking out."}
        arbSubTitle={"أضف بعض المنتجات إلى السلة قبل إتمام الشراء."}
      />
    );
  else
    body = (
      <div className={gridContainerClass}>
        <div className={leftSectionClass}>
          <OrderData selectedStage={selectedStage} setSelectedStage={setSelectedStage} orderStages={orderStages} />
          <div className={FormContainerClass}>
            <div>
              <Step selectedStage={selectedStage} setSelectedStage={setSelectedStage} />
            </div>
          </div>
        </div>
        <div className={rightSectionClass}>
          <OrderSummery />
        </div>
      </div>
    );

  return (
    <div className={`${PurchaseConfirmationContianerClass} ${lang === "arb" ? arbPurchaseConfirmationContianerClass : ""}`}>
      <UserSettings />
      <SearchBar value={1} />
      <PurchaseConfirmationHeader />
      {body}
    </div>
  );
}

export default PurchaseConfirmation;
