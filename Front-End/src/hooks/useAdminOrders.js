import { useCallback, useEffect, useState } from "react";
import { fetchAllOrders } from "../api/admin";

// All orders for the admin panel (Orders page and Dashboard).
export function useAdminOrders() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);

  const reload = useCallback(async () => {
    try {
      setOrders(await fetchAllOrders());
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    reload().catch(() => {});
  }, [reload]);

  return { orders, loading, reload };
}
