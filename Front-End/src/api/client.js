// Minimal fetch wrapper for the Go backend.
const BASE_URL = (import.meta.env.VITE_API_URL || "http://localhost:8080").replace(/\/$/, "");
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

export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

let onUnauthorized = null;
// Called when a request that carried a token comes back 401 (expired/invalid session).
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

export async function api(path, { method = "GET", body } = {}) {
  const token = tokenStore.get();
  const headers = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (token) headers.Authorization = `Bearer ${token}`;

  let response;
  try {
    response = await fetch(BASE_URL + path, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
    throw new ApiError(0, "Cannot reach the server. Is the backend running?");
  }

  const text = await response.text();
  const data = text ? parse(text) : null;

  if (!response.ok) {
    if (response.status === 401 && token) onUnauthorized?.();
    const message = typeof data === "string" ? data.trim() : data?.message;
    throw new ApiError(response.status, message || response.statusText);
  }

  return data;
}
