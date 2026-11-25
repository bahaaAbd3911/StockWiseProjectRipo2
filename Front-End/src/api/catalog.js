import { api, fetchAll, upload } from "./client";
import { fromCategory, fromProduct, toCategory, toProduct, toReview } from "./mappers";

// The whole catalog (the shop filters it in the browser). Admins also get
// hidden products, so they can edit and re-activate them.
export const fetchProducts = async (admin = false) =>
  (await fetchAll(admin ? "/api/admin/products" : "/api/products")).map(toProduct);

// Admins also get switched-off categories.
export const fetchCategories = async (admin = false) =>
  (await api(admin ? "/api/admin/categories" : "/api/categories")).map(toCategory);

export const createProduct = async (p) =>
  toProduct(await api("/api/products", { method: "POST", body: fromProduct(p) }));
export const updateProduct = async (id, p) =>
  toProduct(await api(`/api/products/${id}`, { method: "PUT", body: fromProduct(p) }));
// Hides the product (kept for order history).
export const deleteProduct = (id) => api(`/api/products/${id}`, { method: "DELETE" });

export const createCategory = async (c) =>
  toCategory(await api("/api/admin/categories", { method: "POST", body: fromCategory(c) }));
export const updateCategory = async (id, c) =>
  toCategory(await api(`/api/admin/categories/${id}`, { method: "PUT", body: fromCategory(c) }));
// Removes the category and hides its products.
export const deleteCategory = (id) => api(`/api/admin/categories/${id}`, { method: "DELETE" });

export const fetchReviews = async (productId) =>
  (await api(`/api/products/${productId}/reviews`)).map(toReview);
export const postReview = async (productId, comment, rating) =>
  toReview(await api(`/api/products/${productId}/reviews`, { method: "POST", body: { comment, rating } }));

export const fetchSettings = () => api("/api/settings");
export const saveSettings = (settings) => api("/api/admin/settings", { method: "PUT", body: settings });

// Uploads a banner image for this store's settings and returns its URL.
export const uploadSettingsImage = async (file) => (await upload("/api/admin/settings/images", file)).url;
