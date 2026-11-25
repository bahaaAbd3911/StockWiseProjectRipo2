import { api } from "./client";
import { splitName, toFavoriteProduct, toOrder, toUser } from "./mappers";

// The admin form uses one "full name" field; the API stores first/last names.
const fromUserForm = (form) => ({
  ...splitName(form.userFullName),
  email: form.userEmail,
  phone: form.phone ?? "",
  password: form.userPassword ?? "",
  role: form.role,
  is_active: form.status !== "inactive",
});

export const fetchUsers = async () => (await api("/api/admin/users")).map(toUser);
export const fetchUserDetails = async (id) => {
  const data = await api(`/api/admin/users/${id}`);
  return {
    user: toUser(data.user),
    orders: (data.orders ?? []).map(toOrder),
    favorites: (data.favorites ?? []).map(toFavoriteProduct),
  };
};
export const createUser = async (form) =>
  toUser(await api("/api/admin/users", { method: "POST", body: fromUserForm(form) }));
export const updateUser = async (id, form) =>
  toUser(await api(`/api/admin/users/${id}`, { method: "PUT", body: fromUserForm(form) }));
export const deleteUser = (id) => api(`/api/admin/users/${id}`, { method: "DELETE" });

export const fetchAllOrders = async () => (await api("/api/admin/orders")).map(toOrder);
export const updateOrderStatus = (id, status) =>
  api(`/api/admin/orders/${id}/status`, { method: "PUT", body: { status } });
