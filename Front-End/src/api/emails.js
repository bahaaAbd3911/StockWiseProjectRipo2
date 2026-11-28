import { api } from "./client";

// The store's email queue (backend: internal/notify, /api/admin/emails).

const toEmail = (e) => ({
  id: e.id,
  kind: e.kind,
  to: e.to,
  subject: e.subject,
  body: e.body ?? "",
  status: e.status, // pending | failed | sent
  attempts: e.attempts,
  lastError: e.last_error ?? "",
  nextAttemptAt: e.next_attempt_at,
  sentAt: e.sent_at,
  createdAt: e.created_at,
});

export const fetchEmailSummary = async () => {
  const s = await api("/api/admin/emails/summary");
  return { pending: s.pending, failed: s.failed, sentLast24h: s.sent_last_24h, oldestPendingAt: s.oldest_pending_at };
};

// One page of the list: { items, total }.
export const fetchEmails = async ({ status = "", kind = "", to = "", limit = 20, offset = 0 } = {}) => {
  const q = new URLSearchParams({ limit, offset });
  if (status) q.set("status", status);
  if (kind) q.set("kind", kind);
  if (to.trim()) q.set("to", to.trim());
  const res = await api(`/api/admin/emails?${q}`);
  return { items: (res.items ?? []).map(toEmail), total: res.total ?? 0 };
};

export const fetchEmail = async (id) => toEmail(await api(`/api/admin/emails/${id}`));
export const retryEmail = (id) => api(`/api/admin/emails/${id}/retry`, { method: "POST" });
export const retryFailedEmails = () => api("/api/admin/emails/retry-failed", { method: "POST" });
export const cancelEmail = (id) => api(`/api/admin/emails/${id}`, { method: "DELETE" });
