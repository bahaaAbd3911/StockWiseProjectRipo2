import { useCallback, useEffect, useState } from "react";
import * as shopApi from "../api/shop";

const EMPTY = { cart: [], favorites: [], orders: [] };

// Loads the signed-in user's cart, favorites and orders. When signed out it returns
// empty lists (derived, so nothing from the previous user can leak through).
export function useShopData(isSign) {
  const [data, setData] = useState(EMPTY);

  const refresh = useCallback(async (key, load) => {
    const value = await load();
    setData((prev) => ({ ...prev, [key]: value }));
  }, []);

  const refreshCart = useCallback(() => refresh("cart", shopApi.fetchCart), [refresh]);
  const refreshFavorites = useCallback(() => refresh("favorites", shopApi.fetchFavorites), [refresh]);
  const refreshOrders = useCallback(() => refresh("orders", shopApi.fetchMyOrders), [refresh]);

  useEffect(() => {
    if (!isSign) return;
    Promise.all([shopApi.fetchCart(), shopApi.fetchFavorites(), shopApi.fetchMyOrders()])
      .then(([cart, favorites, orders]) => setData({ cart, favorites, orders }))
      .catch(() => {});
  }, [isSign]);

  return { ...(isSign ? data : EMPTY), refreshCart, refreshFavorites, refreshOrders };
}
