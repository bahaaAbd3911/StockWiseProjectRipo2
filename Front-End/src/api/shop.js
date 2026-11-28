import { api } from "./client";
import { toCartBundle, toCartItem, toFavoriteProduct, toOrder } from "./mappers";

// The cart: product lines, bundle lines, and what the backend worked out -
// subtotal, the coupon (with a "problem" when it stopped working), discount,
// shipping and total. city: the delivery city to estimate shipping for
// (empty = the customer's default address).
export const fetchCart = async (city = "") => {
  const c = await api(city ? `/api/cart?city=${encodeURIComponent(city)}` : "/api/cart");
  return {
    items: (c.items ?? []).map(toCartItem),
    bundles: (c.bundles ?? []).map(toCartBundle),
    subtotal: Number(c.subtotal) || 0,
    discount: Number(c.discount) || 0,
    shipping: Number(c.shipping) || 0,
    shippingCity: c.shipping_city ?? "",
    freeShippingOver: c.free_shipping_over == null ? null : Number(c.free_shipping_over),
    total: Number(c.total) || 0,
    coupon: c.coupon
      ? { code: c.coupon.code, discount: Number(c.coupon.discount_amount) || 0, problem: c.coupon.problem ?? "" }
      : null,
  };
};
export const addCartItem = (productId, quantity, color) =>
  api("/api/cart/items", {
    method: "POST",
    body: { product_id: productId, quantity, selected_color: color || null },
  });
export const updateCartItem = (itemId, quantity) =>
  api(`/api/cart/items/${itemId}`, { method: "PUT", body: { quantity } });
export const removeCartItem = (itemId) => api(`/api/cart/items/${itemId}`, { method: "DELETE" });

export const fetchFavorites = async () => ((await api("/api/favorites")) ?? []).map(toFavoriteProduct);
export const addFavorite = (productId) =>
  api("/api/favorites", { method: "POST", body: { product_id: productId } });
export const removeFavorite = (productId) => api(`/api/favorites/${productId}`, { method: "DELETE" });

export const createAddress = (address) => api("/api/addresses", { method: "POST", body: address });

export const fetchMyOrders = async () => ((await api("/api/orders")) ?? []).map(toOrder);
export const fetchMyOrder = async (id) => toOrder(await api(`/api/orders/${id}`));
export const placeOrder = async (addressId, paymentMethod) =>
  toOrder(await api("/api/orders", { method: "POST", body: { address_id: addressId, payment_method: paymentMethod } }));
export const cancelMyOrder = (orderId) => api(`/api/orders/${orderId}/cancel`, { method: "POST" });
