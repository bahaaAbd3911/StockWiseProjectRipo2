import { api } from "./client";

// The platform super admin's API (served only on the platform host, e.g.
// admin.myplatform.com). Its login is stored separately from shop logins.
const KEY = "platformToken";

export const platformToken = {
  get: () => {
    try {
      return localStorage.getItem(KEY) ?? "";
    } catch {
      return "";
    }
  },
  set: (t) => localStorage.setItem(KEY, t),
  clear: () => localStorage.removeItem(KEY),
};

const call = (path, options = {}) => api(path, { ...options, token: platformToken.get() });

export const platformLogin = async (email, password) => {
  const res = await api("/api/platform/login", { method: "POST", body: { email, password }, token: "" });
  platformToken.set(res.token);
  return res.admin;
};
export const platformMe = () => call("/api/platform/me");

export const listStores = ({ search = "", status = "", offset = 0, limit = 20 } = {}) => {
  const q = new URLSearchParams({ limit, offset });
  if (search) q.set("search", search);
  if (status) q.set("status", status);
  return call(`/api/platform/stores?${q}`);
};
// body: {slug, name, currency, timezone, admin: {first_name, last_name, email}}
export const createStore = (body) => call("/api/platform/stores", { method: "POST", body });
export const updateStore = (id, body) => call(`/api/platform/stores/${id}`, { method: "PUT", body });
export const suspendStore = (id) => call(`/api/platform/stores/${id}/suspend`, { method: "POST" });
export const activateStore = (id) => call(`/api/platform/stores/${id}/activate`, { method: "POST" });
export const resendInvite = (id) => call(`/api/platform/stores/${id}/invite`, { method: "POST" });

// Is this page opened on the platform admin host? (VITE_PLATFORM_HOST, or any admin.* host.)
export const isPlatformHost = () => {
  const host = window.location.hostname;
  const configured = import.meta.env.VITE_PLATFORM_HOST;
  return configured ? host === configured : host.startsWith("admin.");
};
