import { api } from "./client";
import { fromCategory, fromProduct, toCategory, toProduct, toReview } from "./mappers";

export const fetchProducts = async () => (await api("/api/products")).map(toProduct);
export const fetchCategories = async () => (await api("/api/categories")).map(toCategory);

export const createProduct = async (p) =>
  toProduct(await api("/api/products", { method: "POST", body: fromProduct(p) }));
export const updateProduct = async (id, p) =>
  toProduct(await api(`/api/products/${id}`, { method: "PUT", body: fromProduct(p) }));
export const deleteProduct = (id) => api(`/api/products/${id}`, { method: "DELETE" });

export const createCategory = async (c) =>
  toCategory(await api("/api/admin/categories", { method: "POST", body: fromCategory(c) }));
export const updateCategory = async (id, c) =>
  toCategory(await api(`/api/admin/categories/${id}`, { method: "PUT", body: fromCategory(c) }));
export const deleteCategory = (id) => api(`/api/admin/categories/${id}`, { method: "DELETE" });

export const fetchReviews = async (productId) =>
  (await api(`/api/products/${productId}/reviews`)).map(toReview);
export const postReview = async (productId, comment, rating) =>
  toReview(await api(`/api/products/${productId}/reviews`, { method: "POST", body: { comment, rating } }));

export const fetchSettings = () => api("/api/settings");
export const saveSettings = (settings) => api("/api/admin/settings", { method: "PUT", body: settings });
