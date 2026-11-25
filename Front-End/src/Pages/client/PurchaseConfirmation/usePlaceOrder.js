import { useContext, useState } from "react";
import { useNavigate } from "react-router";

import { AuthContext } from "../../../context/AuthContext";
import { useShop } from "../../../hooks/useShop";
import { useToast } from "../../../hooks/useToast";
import { submitOrderValidation } from "../../../utils/helpers";

// Checkout payment titles (PaymentMethod.jsx) -> backend payment_method values.
export const PAYMENT_KEYS = {
  "Cash On Delivery": "cash_on_delivery",
  "Online Payment": "credit_card",
  PayPal: "paypal",
};

// Field indexes: 0 name, 1 phone, 2 email, 3 city, 4 area, 5 address, 6 delivery, 7 payment.
const SHIPPING_FIELDS = [0, 1, 2, 3, 4, 5, 6];

export function usePlaceOrder() {
  const { orderFormFields, setorderFormFields, setOrderStages, resetOrderForm, openProfileTab } = useContext(AuthContext);
  const { checkout } = useShop();
  const toast = useToast();
  const navigate = useNavigate();
  const [placing, setPlacing] = useState(false);

  // Validates every step and flags the stages that still have errors.
  const validate = () => {
    const ok = submitOrderValidation(orderFormFields, setorderFormFields);
    const shippingError = SHIPPING_FIELDS.some((i) => orderFormFields[i].data === "" || orderFormFields[i].flag);
    const paymentError = !orderFormFields[7].data;
    setOrderStages((prev) =>
      prev.map((stage, i) => ({ ...stage, alert: i === 0 ? !ok && shippingError : i === 1 ? !ok && paymentError : false })),
    );
    return ok;
  };

  const placeOrder = async () => {
    if (placing || !validate()) return;
    const f = orderFormFields.map((field) => String(field.data).trim());

    setPlacing(true);
    try {
      const order = await checkout(
        {
          recipient_name: f[0],
          phone: f[1],
          city: f[3],
          area: f[4],
          street: f[5],
          additional_instructions: `${f[6]} · ${f[2]}`,
          is_default: true,
        },
        PAYMENT_KEYS[f[7]] ?? "cash_on_delivery",
      );
      toast.success(`Order #${order.id} placed successfully!`, `تم تقديم الطلب #${order.id} بنجاح!`);
      resetOrderForm();
      openProfileTab(1);
      navigate("/client/profile");
    } catch (err) {
      toast.error(err.message, err.message);
    } finally {
      setPlacing(false);
    }
  };

  return { placeOrder, placing };
}
