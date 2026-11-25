import { api } from "./client";
import { toCartItem, toFavoriteProduct, toOrder } from "./mappers";

export const fetchCart = async () => ((await api("/api/cart")).items ?? []).map(toCartItem);
export const addCartItem = (productId, quantity, color) =>
  api("/api/cart/items", {
    method: "POST",
    body: { product_id: productId, quantity, selected_color: color || null },
  });
export const updateCartItem = (itemId, quantity) =>
  api(`/api/cart/items/${itemId}`, { method: "PUT", body: { quantity } });
export const removeCartItem = (itemId) => api(`/api/cart/items/${itemId}`, { method: "DELETE" });

export const fetchFavorites = async () => (await api("/api/favorites")).map(toFavoriteProduct);
export const addFavorite = (productId) =>
  api("/api/favorites", { method: "POST", body: { product_id: productId } });
export const removeFavorite = (productId) => api(`/api/favorites/${productId}`, { method: "DELETE" });

export const createAddress = (address) => api("/api/addresses", { method: "POST", body: address });

export const fetchMyOrders = async () => ((await api("/api/orders")) ?? []).map(toOrder);
export const placeOrder = async (addressId, paymentMethod) =>
  toOrder(await api("/api/orders", { method: "POST", body: { address_id: addressId, payment_method: paymentMethod } }));
export const cancelMyOrder = (orderId) => api(`/api/orders/${orderId}/cancel`, { method: "POST" });
