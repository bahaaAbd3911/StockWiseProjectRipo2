// Minimal fetch wrapper for the Go backend.
// By default the API is on the same address as the shop (/api): the backend
// knows which store this is from the host name. VITE_API_URL can point
// elsewhere (single-store setups with the API on another domain).
const BASE_URL = (import.meta.env.VITE_API_URL || "").replace(/\/$/, "");
const TOKEN_KEY = "authToken";

export const tokenStore = {
  get: () => {
    try {
      return localStorage.getItem(TOKEN_KEY);
    } catch {
      return null;
    }
  },
  set: (token) => localStorage.setItem(TOKEN_KEY, token),
  clear: () => localStorage.removeItem(TOKEN_KEY),
};

// status: HTTP status (0 = server unreachable). fields: per-field messages from
// a validation error, e.g. { email: "is required" }. retryAfter: seconds to wait
// after a 429.
export class ApiError extends Error {
  constructor(status, message, fields = {}, retryAfter = 0) {
    super(message);
    this.status = status;
    this.fields = fields;
    this.retryAfter = retryAfter;
  }
}

let onUnauthorized = null;
// Called when a request that carried a token comes back 401 (expired/invalid/revoked session).
export const setUnauthorizedHandler = (handler) => {
  onUnauthorized = handler;
};

const parse = (text) => {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
};

// The backend answers errors as {"error": "...", "fields": {"email": "is required"}}.
// "validation failed" alone says little, so the first field message is shown with it.
const errorMessage = (data, fallback) => {
  if (typeof data === "string") return data.trim() || fallback;
  const fields = data?.fields ?? {};
  const [field, msg] = Object.entries(fields)[0] ?? [];
  if (field) return `${field.replace(/_/g, " ")}: ${msg}`;
  return data?.error || data?.message || fallback;
};

// token: send this token instead of the shop login (platform admin calls).
// The "logged out" handler only runs for the shop login.
async function send(path, { method = "GET", body, form, token: explicitToken } = {}) {
  const token = explicitToken ?? tokenStore.get();
  const headers = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (token) headers.Authorization = `Bearer ${token}`;

  let response;
  try {
    response = await fetch(BASE_URL + path, {
      method,
      headers,
      // FormData sets its own multipart Content-Type (with the boundary).
      body: form ?? (body !== undefined ? JSON.stringify(body) : undefined),
    });
  } catch {
    throw new ApiError(0, "Cannot reach the server. Is the backend running?");
  }

  const text = await response.text();
  const data = text ? parse(text) : null;

  if (!response.ok) {
    if (response.status === 401 && token && explicitToken === undefined) onUnauthorized?.();
    throw new ApiError(
      response.status,
      errorMessage(data, response.statusText),
      data?.fields ?? {},
      Number(response.headers.get("Retry-After")) || 0,
    );
  }
  return data;
}

export const api = (path, options) => send(path, options);

// Uploads one file as multipart/form-data in the field "image" (plus text fields).
export function upload(path, file, fields = {}, method = "POST") {
  const form = new FormData();
  form.append("image", file);
  Object.entries(fields).forEach(([key, value]) => form.append(key, String(value)));
  return send(path, { method, form });
}

// Lists come back one page at a time: {items, total, limit, offset}. This reads
// every page, for screens that work on the whole list (catalog, admin tables).
export async function fetchAll(path, pageSize = 100) {
  const sep = path.includes("?") ? "&" : "?";
  const items = [];
  for (let offset = 0; ; offset += pageSize) {
    const page = await send(`${path}${sep}limit=${pageSize}&offset=${offset}`);
    items.push(...(page?.items ?? []));
    if (!page || items.length >= page.total || !page.items?.length) return items;
  }
}

// Turns a data URL (what FileReader gives) back into a file for upload.
export const dataUrlToFile = async (dataUrl, name = "image") => {
  const blob = await (await fetch(dataUrl)).blob();
  const ext = (blob.type.split("/")[1] || "png").replace("jpeg", "jpg");
  return new File([blob], `${name}.${ext}`, { type: blob.type });
};

export const isDataUrl = (value) => typeof value === "string" && value.startsWith("data:");
