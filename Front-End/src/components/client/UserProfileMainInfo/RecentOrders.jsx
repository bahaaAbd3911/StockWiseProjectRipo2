import { useContext } from "react";

import {
  RecentOrdersContainerClass,
  RecentOrdersHeaderClass,
  ordersListClass,
} from "./UserProfileMainInfo.module.css";
import { AuthContext } from "../../../context/AuthContext";
import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import Order from "./Order";

function RecentOrders() {
  const { openProfileTab } = useContext(AuthContext);
  const { orders } = useShop();
  const { tr } = useLang();

  return (
    <div className={RecentOrdersContainerClass}>
      <div className={RecentOrdersHeaderClass}>
        <div>
          <div>
            <i className="fa-solid fa-clock"></i>
          </div>
          <div>{tr("Recent Orders", "الطلبات الأخيرة")}</div>
        </div>
        <div onClick={() => openProfileTab(1)} style={{ cursor: "pointer" }}>
          <div>{tr("View All", "عرض الكل")}</div>
          <div>
            <i className="fa-solid fa-arrow-right-long"></i>
          </div>
        </div>
      </div>
      <div className={ordersListClass}>
        {orders.length === 0 && <p>{tr("You haven't placed any orders yet.", "لم تقم بأي طلب بعد.")}</p>}
        {orders.slice(0, 3).map((order) => (
          <Order data={order} key={order.id} />
        ))}
      </div>
    </div>
  );
}

export default RecentOrders;
