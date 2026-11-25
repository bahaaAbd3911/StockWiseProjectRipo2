import { useEffect, useState } from "react";

import { fetchUsers } from "../../../api/admin";
import { useAdminOrders } from "../../../hooks/useAdminOrders";
import { useCatalog } from "../../../hooks/useCatalog";
import { buildStats, recentActivity, topProducts } from "./dashboardStats";

// Everything the dashboard shows, computed from live API data.
export function useDashboardData() {
  const { products } = useCatalog();
  const { orders } = useAdminOrders();
  const [users, setUsers] = useState([]);

  useEffect(() => {
    fetchUsers().then(setUsers).catch(() => {});
  }, []);

  return {
    orders,
    stats: buildStats({ users, orders, products }),
    recentOrders: orders.slice(0, 5),
    recentUsers: users.slice(0, 4),
    topSellers: topProducts(orders, products),
    activities: recentActivity(orders, users),
  };
}
