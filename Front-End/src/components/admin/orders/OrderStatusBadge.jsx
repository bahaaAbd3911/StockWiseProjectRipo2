import { StatusBadge } from "../ui";
import { ORDER_STATUS_LABELS, ORDER_STATUS_TONES } from "../../../constants/orderStatus";
import { pickLabel } from "../../../utils/format";

export default function OrderStatusBadge({ status, isAr }) {
  return <StatusBadge tone={ORDER_STATUS_TONES[status]}>{pickLabel(ORDER_STATUS_LABELS[status], isAr)}</StatusBadge>;
}
