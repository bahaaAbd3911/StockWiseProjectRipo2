import { api } from "./client";

// The store's custom domains (store admin).
export const fetchDomains = () => api("/api/admin/domains");
export const addDomain = (domain) => api("/api/admin/domains", { method: "POST", body: { domain } });
export const verifyDomain = (id) => api(`/api/admin/domains/${id}/verify`, { method: "POST" });
export const makePrimaryDomain = (id) => api(`/api/admin/domains/${id}/primary`, { method: "POST" });
export const removeDomain = (id) => api(`/api/admin/domains/${id}`, { method: "DELETE" });
