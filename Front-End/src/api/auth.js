import { api } from "./client";

export const login = (email, password) =>
  api("/api/login", { method: "POST", body: { email, password } });

// Public sign-up: the backend always creates a customer.
export const register = (data) => api("/api/register", { method: "POST", body: data });

export const getProfile = () => api("/api/profile");

export const updateProfile = (data) => api("/api/profile", { method: "PUT", body: data });

export const changePassword = (currentPassword, newPassword) =>
  api("/api/profile/password", {
    method: "PUT",
    body: { current_password: currentPassword, new_password: newPassword },
  });
