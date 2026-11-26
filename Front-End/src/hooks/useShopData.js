import { useCallback, useEffect, useState } from "react";
import * as shopApi from "../api/shop";

const EMPTY_CART_INFO = { bundles: [], subtotal: 0, discount: 0, total: 0, coupon: null };
const EMPTY = { cart: [], cartInfo: EMPTY_CART_INFO, favorites: [], orders: [] };

// cart = the product lines (what most components use); cartInfo = bundle lines,
// the coupon and the totals the backend worked out.
const splitCart = ({ items, ...info }) => ({ cart: items, cartInfo: info });

// Loads the signed-in user's cart, favorites and orders. When signed out it returns
// empty lists (derived, so nothing from the previous user can leak through).
export function useShopData(isSign) {
  const [data, setData] = useState(EMPTY);

  const refreshCart = useCallback(async () => {
    const cart = splitCart(await shopApi.fetchCart());
    setData((prev) => ({ ...prev, ...cart }));
  }, []);
  const refreshFavorites = useCallback(async () => {
    const favorites = await shopApi.fetchFavorites();
    setData((prev) => ({ ...prev, favorites }));
  }, []);
  const refreshOrders = useCallback(async () => {
    const orders = await shopApi.fetchMyOrders();
    setData((prev) => ({ ...prev, orders }));
  }, []);

  useEffect(() => {
    if (!isSign) return;
    Promise.all([shopApi.fetchCart(), shopApi.fetchFavorites(), shopApi.fetchMyOrders()])
      .then(([cart, favorites, orders]) => setData({ ...splitCart(cart), favorites, orders }))
      .catch(() => {});
  }, [isSign]);

  return { ...(isSign ? data : EMPTY), refreshCart, refreshFavorites, refreshOrders };
}
