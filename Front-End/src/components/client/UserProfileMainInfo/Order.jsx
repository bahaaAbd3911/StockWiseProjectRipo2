import {
  RecentOrdersDataClass,
  imgContainerClass,
  waitingClass,
  shippingClass,
  deliveredClass,
} from "./UserProfileMainInfo.module.css";
import { useLang } from "../../../hooks/useLang";
import { ORDER_STATUS_LABELS } from "../../../constants/orderStatus";
import { formatDate, pickLabel } from "../../../utils/format";

const STATUS_CLASS = {
  pending: waitingClass,
  confirmed: waitingClass,
  processing: shippingClass,
  shipped: shippingClass,
  delivered: deliveredClass,
  cancelled: waitingClass,
};

// One row in the profile's "Recent Orders" box.
function Order({ data }) {
  const { tr, isAr } = useLang();

  return (
    <div className={RecentOrdersDataClass}>
      <div>
        <div className={imgContainerClass}>
          <img src={data.items[0]?.image} alt="" />
        </div>
        <div>
          <div>
            {tr("Order", "الطلب")} #{data.id}
          </div>
          <div>{formatDate(data.createdAt, isAr)}</div>
        </div>
      </div>
      <div className={STATUS_CLASS[data.status]}>{pickLabel(ORDER_STATUS_LABELS[data.status], isAr)}</div>
    </div>
  );
}

export default Order;
