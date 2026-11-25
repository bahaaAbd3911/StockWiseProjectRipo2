import { useContext } from "react";

import { ShopContext } from "./ShopContext";
import { AuthContext } from "./AuthContext";
import { CatalogContext } from "./CatalogContext";
import * as shopApi from "../api/shop";
import { useShopData } from "../hooks/useShopData";
import { useToast } from "../hooks/useToast";

export function ShopProvider({ children }) {
  const { isSign } = useContext(AuthContext);
  const { reload: reloadCatalog } = useContext(CatalogContext);
  const toast = useToast();
  const data = useShopData(isSign);

  // Every cart/favorite action goes through this guard.
  const requireSignIn = () => {
    if (isSign) return true;
    toast.error("You must sign in first", "يجب عليك تسجيل الدخول أولاً");
    return false;
  };

  // Runs an API call and shows the server's message if it fails.
  const run = async (action, successEn, successAr) => {
    try {
      await action();
      if (successEn) toast.success(successEn, successAr);
      return true;
    } catch (err) {
      toast.error(err.message, err.message);
      return false;
    }
  };

  const cartItemFor = (productId) => data.cart.find((item) => item.productID === productId);
  const isFavorite = (productId) => data.favorites.some((p) => p.id === productId);

  const addToCart = (product, quantity = 1, color = null) => {
    if (!requireSignIn()) return Promise.resolve(false);
    if (product.outOfStock) {
      toast.error("This product is out of stock", "نفد هذا المنتج من المخزون");
      return Promise.resolve(false);
    }
    const defaultColor = color ?? product.availableColors?.[0]?.name ?? null;
    return run(
      async () => {
        await shopApi.addCartItem(product.id, quantity, defaultColor);
        await data.refreshCart();
      },
      "Product added to your cart",
      "تمت إضافة المنتج إلى سلة التسوق",
    );
  };

  const updateCartQuantity = (itemId, quantity) =>
    run(async () => {
      await shopApi.updateCartItem(itemId, quantity);
      await data.refreshCart();
    });

  const removeFromCart = (itemId) =>
    run(
      async () => {
        await shopApi.removeCartItem(itemId);
        await data.refreshCart();
      },
      "Removed from your cart",
      "تمت الإزالة من السلة",
    );

  const toggleFavorite = (product) => {
    if (!requireSignIn()) return Promise.resolve(false);
    const remove = isFavorite(product.id);
    return run(
      async () => {
        await (remove ? shopApi.removeFavorite(product.id) : shopApi.addFavorite(product.id));
        await data.refreshFavorites();
      },
      remove ? "Removed from your favorites" : "Added to your favorites",
      remove ? "تمت الإزالة من المفضلة" : "تمت الإضافة إلى المفضلة",
    );
  };

  // Saves the delivery address, places the order, then refreshes cart, orders and stock.
  const checkout = async (address, paymentMethod) => {
    const saved = await shopApi.createAddress(address);
    const order = await shopApi.placeOrder(saved.id, paymentMethod);
    await Promise.all([data.refreshCart(), data.refreshOrders(), reloadCatalog()]);
    return order;
  };

  const cancelOrder = (orderId) =>
    run(
      async () => {
        await shopApi.cancelMyOrder(orderId);
        await Promise.all([data.refreshOrders(), reloadCatalog()]);
      },
      "Order cancelled",
      "تم إلغاء الطلب",
    );

  return (
    <ShopContext.Provider
      value={{
        ...data,
        cartCount: data.cart.reduce((sum, item) => sum + item.quantity, 0),
        requireSignIn,
        cartItemFor,
        isFavorite,
        addToCart,
        updateCartQuantity,
        removeFromCart,
        toggleFavorite,
        checkout,
        cancelOrder,
      }}
    >
      {children}
    </ShopContext.Provider>
  );
}
