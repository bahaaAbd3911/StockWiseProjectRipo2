import { PAYMENT_METHODS } from "../../../constants/paymentMethods";
import { pickLabel } from "../../../utils/format";

export default function PaymentMethod({ method, isAr }) {
  const info = PAYMENT_METHODS[method];
  if (!info) return method;

  return (
    <span style={{ display: "inline-flex", alignItems: "center", gap: 8 }}>
      <i className={info.icon} style={{ color: "var(--adm-green)" }} />
      {pickLabel(info, isAr)}
    </span>
  );
}
