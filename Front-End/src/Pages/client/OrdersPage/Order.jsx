import { useState } from "react";

import {
  OrderContainerClass,
  orderLogoClass,
  orderDataClass,
  orderImagesClass,
  orderPriceClass,
  orderViewDetailsBtnClass,
  orderImgClass,
  orderCardClass,
} from "./ClientOrdersPage.module.css";
import { useLang } from "../../../hooks/useLang";
import { ORDER_STATUS_ICONS, ORDER_STATUS_LABELS } from "../../../constants/orderStatus";
import { formatDate, formatMoney, orderTotal, pickLabel } from "../../../utils/format";
import { STATUS_STYLE } from "./orderGroups";
import OrderDetails from "./OrderDetails";

function Order({ order }) {
  const { tr, isAr } = useLang();
  const [open, setOpen] = useState(false);
  const itemCount = order.items.reduce((sum, i) => sum + i.quantity, 0);

  return (
    <div className={`${orderCardClass} ${STATUS_STYLE[order.status]}`}>
      <div className={OrderContainerClass}>
        <div className={orderLogoClass}>
          <i className={ORDER_STATUS_ICONS[order.status]}></i>
        </div>
        <div className={orderDataClass}>
          <div>
            {tr("Order", "الطلب")} #{order.id}
          </div>
          <div>{formatDate(order.createdAt, isAr)}</div>
          <div>{pickLabel(ORDER_STATUS_LABELS[order.status], isAr)}</div>
        </div>
        <div className={orderImagesClass}>
          {order.items.slice(0, 3).map((item) => (
            <div className={orderImgClass} key={item.id}>
              <img src={item.image} alt="" />
            </div>
          ))}
        </div>
        <div className={orderPriceClass}>
          <div>{formatMoney(orderTotal(order))}</div>
          <div>
            {itemCount} {tr("Items", "عناصر")}
          </div>
        </div>
        <button
          type="button"
          className={orderViewDetailsBtnClass}
          aria-expanded={open}
          onClick={() => setOpen((v) => !v)}
        >
          <i className="fa-solid fa-eye"></i>
          <span>{open ? tr("Hide Details", "إخفاء التفاصيل") : tr("View Details", "عرض التفاصيل")}</span>
          <i className={`fa-solid ${open ? "fa-angle-up" : "fa-angle-down"}`}></i>
        </button>
      </div>
      {open && <OrderDetails order={order} />}
    </div>
  );
}

export default Order;
