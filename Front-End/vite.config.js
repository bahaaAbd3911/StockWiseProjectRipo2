import process from "node:process";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The app calls the API at /api on its own address, because the backend
// picks the store from the host name (ahmad.localhost:5173 -> store "ahmad").
// In development this proxy passes /api and /uploads to the Go backend and
// keeps the original Host header (changeOrigin: false), so the backend sees
// which store was opened. In production the same paths are routed to the API
// by Caddy / the hosting platform.
const backend = process.env.VITE_BACKEND || "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    // *.localhost names (ahmad.localhost, admin.localhost) resolve to this machine.
    allowedHosts: [".localhost"],
    proxy: {
      "/api": { target: backend, changeOrigin: false },
      "/uploads": { target: backend, changeOrigin: false },
    },
  },
});
