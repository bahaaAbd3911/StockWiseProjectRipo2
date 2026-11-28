import { useLang } from "../../../hooks/useLang";
import { formatMoney } from "../../../utils/format";

// The shipping row of a cart or checkout summary, worked out by the backend
// (cartInfo.shipping for cartInfo.shippingCity). Also tells the customer how
// much more ships free, when the store has a "free over" amount.
// rowClass / hintClass: the summary's own row and note styles; Cell: the
// element its rows use for the two cells ("span" or "div").
export default function ShippingLine({
  cartInfo,
  rowClass,
  hintClass,
  label,
  cell: Cell = "span",
}) {
  const { tr } = useLang();
  const { shipping, shippingCity, freeShippingOver, subtotal, discount } =
    cartInfo;
  const itemsTotal = subtotal - discount;
  const missing =
    freeShippingOver != null && shipping > 0
      ? freeShippingOver - itemsTotal
      : 0;

  return (
    <>
      <div className={rowClass}>
        <Cell>
          {label ?? tr("Shipping", "الشحن")}
          {shippingCity && shipping > 0 ? ` (${shippingCity})` : ""}
        </Cell>
        <Cell>
          {shipping > 0 ? formatMoney(shipping) : tr("Free", "مجاني")}
        </Cell>
      </div>
      {missing > 0 && (
        <small className={hintClass}>
          {tr(
            `Add ${formatMoney(missing)} more for free shipping`,
            `أضف ${formatMoney(missing)} للحصول على شحن مجاني`,
          )}
        </small>
      )}
    </>
  );
}
