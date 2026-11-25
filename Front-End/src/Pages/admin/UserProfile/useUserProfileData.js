import { useCallback, useEffect, useState } from "react";
import { fetchUserDetails } from "../../../api/admin";

const EMPTY = { user: null, orders: [], favorites: [] };

// Loads one user's profile, orders and favorites for the admin profile page.
export function useUserProfileData(userId) {
  const [state, setState] = useState({ id: null, data: EMPTY, status: "loading" });

  const load = useCallback(
    () =>
      fetchUserDetails(userId).then(
        (data) => setState({ id: userId, data, status: "ready" }),
        () => setState({ id: userId, data: EMPTY, status: "missing" }),
      ),
    [userId],
  );

  useEffect(() => {
    load();
  }, [load]);

  // While a different user's data is still showing, report "loading".
  const current = state.id === userId ? state : { data: EMPTY, status: "loading" };

  return { ...current.data, userOrders: current.data.orders, status: current.status, reload: load };
}
