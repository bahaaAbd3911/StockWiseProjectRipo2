import { useState } from "react";

import {
  ClientOrdersPageContainerClass,
  arbClientOrdersPageContainerClass,
  ordersListClass,
} from "./ClientOrdersPage.module.css";
import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import UserOrdersMainInfo from "./UserOrdersMainInfo";
import UserOrdersFilter from "./UserOrdersFilter";
import Order from "./Order";
import { ORDER_GROUPS, inGroup } from "./orderGroups";

// "My Orders" tab — the signed-in user's orders from the database.
function ClientOrdersPage() {
  const { orders } = useShop();
  const { tr, isAr } = useLang();
  const [selected, setSelected] = useState("all");

  const group = ORDER_GROUPS.find((g) => g.key === selected);
  const visible = orders.filter((o) => inGroup(o, group));

  return (
    <div className={`${ClientOrdersPageContainerClass} ${isAr ? arbClientOrdersPageContainerClass : ""}`}>
      <UserOrdersMainInfo />
      <UserOrdersFilter orders={orders} selected={selected} onSelect={setSelected} />
      <div className={ordersListClass}>
        {visible.length === 0 && <p style={{ textAlign: "center", color: "var(--color-text-muted)" }}>{tr("No orders here yet.", "لا توجد طلبات هنا بعد.")}</p>}
        {visible.map((order) => (
          <Order order={order} key={order.id} />
        ))}
      </div>
    </div>
  );
}

export default ClientOrdersPage;
