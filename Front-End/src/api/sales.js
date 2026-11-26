import { api, fetchAll } from "./client";
import { toProduct } from "./mappers";

const num = (v) => Number(v) || 0;

// ---------- shapes ----------

export const toOffer = (o) => ({
  id: o.id,
  name: o.name,
  nameAr: o.name_ar,
  description: o.description ?? "",
  descriptionAr: o.description_ar ?? "",
  discountType: o.discount_type, // percentage | fixed
  discountValue: num(o.discount_value),
  scope: o.scope, // products | category
  category: o.category ?? "",
  productIds: o.product_ids ?? [],
  startsAt: o.starts_at,
  endsAt: o.ends_at,
  isActive: o.is_active,
  status: o.status, // live | scheduled | expired | disabled
});

export const toBundle = (b) => ({
  id: b.id,
  name: b.name,
  nameAr: b.name_ar,
  description: b.description ?? "",
  descriptionAr: b.description_ar ?? "",
  image: b.image_url ?? "",
  price: num(b.price),
  originalTotal: num(b.original_total),
  savings: num(b.savings),
  availableStock: b.available_stock ?? 0,
  isAvailable: b.is_available !== false,
  unavailableReason: b.unavailable_reason ?? "",
  isActive: b.is_active,
  startsAt: b.starts_at,
  endsAt: b.ends_at,
  items: (b.items ?? []).map((i) => ({
    productId: i.product_id,
    name: i.product_name,
    nameAr: i.product_name_ar,
    quantity: i.quantity,
    color: i.selected_color ?? "",
    image: i.primary_image_url ?? "",
  })),
});

export const toCoupon = (c) => ({
  id: c.id,
  code: c.code,
  description: c.description ?? "",
  descriptionAr: c.description_ar ?? "",
  discountType: c.discount_type,
  discountValue: num(c.discount_value),
  scope: c.scope,
  category: c.category ?? "",
  productIds: c.product_ids ?? [],
  minOrder: num(c.min_order_amount),
  maxDiscount: c.max_discount == null ? "" : num(c.max_discount),
  startsAt: c.starts_at,
  endsAt: c.ends_at,
  usageLimit: c.usage_limit ?? "",
  perCustomerLimit: c.per_customer_limit ?? "",
  isActive: c.is_active,
  status: c.status,
  timesUsed: c.times_used ?? 0,
  totalDiscount: num(c.total_discount),
});

export const toDebtAccount = (a) => ({
  userId: a.user_id,
  name: `${a.first_name ?? ""} ${a.last_name ?? ""}`.trim(),
  email: a.email,
  phone: a.phone ?? "",
  creditLimit: num(a.credit_limit),
  balance: num(a.balance),
  availableCredit: num(a.available_credit),
  payLaterAllowed: a.pay_later_allowed,
  lastTransactionAt: a.last_transaction_at,
});

export const toDebtTransaction = (t) => ({
  id: t.id,
  type: t.type, // charge | payment | reversal | adjustment
  amount: num(t.amount),
  balanceAfter: num(t.balance_after),
  note: t.note ?? "",
  orderId: t.order_id,
  createdAt: t.created_at,
});

export const toWaiting = (w) => ({
  id: w.id,
  productId: w.product_id,
  name: w.product_name,
  nameAr: w.product_name_ar,
  stock: w.product_stock,
  quantity: w.quantity,
  color: w.selected_color ?? "",
  status: w.status, // waiting | notified | fulfilled | cancelled
  requestedAt: w.requested_at,
  notifiedAt: w.notified_at,
  customer: w.customer ? `${w.customer.first_name} ${w.customer.last_name}`.trim() : "",
  customerEmail: w.customer?.email ?? "",
});

// ---------- customers ----------

export const fetchLiveOffers = async () => (await fetchAll("/api/offers")).map(toOffer);
export const fetchOfferProducts = async (id) => {
  const res = await api(`/api/offers/${id}`);
  return { offer: toOffer(res.offer), products: (res.products?.items ?? res.products ?? []).map(toProduct) };
};
export const fetchLiveBundles = async () => (await fetchAll("/api/bundles")).map(toBundle);

export const addBundleToCart = (bundleId, quantity = 1) =>
  api("/api/cart/bundles", { method: "POST", body: { bundle_id: bundleId, quantity } });
export const updateCartBundle = (lineId, quantity) =>
  api(`/api/cart/bundles/${lineId}`, { method: "PUT", body: { quantity } });
export const removeCartBundle = (lineId) => api(`/api/cart/bundles/${lineId}`, { method: "DELETE" });

export const applyCoupon = (code) => api("/api/cart/coupon", { method: "POST", body: { code } });
export const removeCoupon = () => api("/api/cart/coupon", { method: "DELETE" });

export const fetchMyDebt = async () => {
  const res = await api("/api/debt");
  return {
    account: toDebtAccount(res.account),
    transactions: (res.transactions?.items ?? []).map(toDebtTransaction),
  };
};

export const joinWaitingList = (productId, quantity = 1, color = null) =>
  api("/api/waiting-list", { method: "POST", body: { product_id: productId, quantity, selected_color: color || null } });
export const fetchMyWaitingList = async () => ((await api("/api/waiting-list")) ?? []).map(toWaiting);
export const cancelWaiting = (id) => api(`/api/waiting-list/${id}`, { method: "DELETE" });

// ---------- store admins ----------

const fromRange = (f) => ({
  starts_at: f.startsAt ? new Date(f.startsAt).toISOString() : null,
  ends_at: f.endsAt ? new Date(f.endsAt).toISOString() : null,
});

const listOrNull = (ids) => (ids ?? []).map(Number).filter(Boolean);

export const fetchAdminOffers = async () => (await fetchAll("/api/admin/offers")).map(toOffer);
const fromOffer = (f) => ({
  name: f.name,
  name_ar: f.nameAr,
  description: f.description,
  description_ar: f.descriptionAr,
  discount_type: f.discountType,
  discount_value: num(f.discountValue),
  scope: f.scope,
  product_ids: f.scope === "products" ? listOrNull(f.productIds) : [],
  category: f.scope === "category" ? f.category : null,
  is_active: f.isActive !== false,
  ...fromRange(f),
});
export const saveOffer = (id, f) =>
  api(id ? `/api/admin/offers/${id}` : "/api/admin/offers", { method: id ? "PUT" : "POST", body: fromOffer(f) });
export const deleteOffer = (id) => api(`/api/admin/offers/${id}`, { method: "DELETE" });

export const fetchAdminBundles = async () => (await fetchAll("/api/admin/bundles")).map(toBundle);
const fromBundle = (f) => ({
  name: f.name,
  name_ar: f.nameAr,
  description: f.description,
  description_ar: f.descriptionAr,
  image_url: f.image || null,
  price: num(f.price),
  is_active: f.isActive !== false,
  ...fromRange(f),
  items: f.items.map((i) => ({ product_id: Number(i.productId), quantity: Number(i.quantity) || 1, selected_color: i.color || null })),
});
export const saveBundle = (id, f) =>
  api(id ? `/api/admin/bundles/${id}` : "/api/admin/bundles", { method: id ? "PUT" : "POST", body: fromBundle(f) });
export const deleteBundle = (id) => api(`/api/admin/bundles/${id}`, { method: "DELETE" });

export const fetchAdminCoupons = async () => (await fetchAll("/api/admin/coupons")).map(toCoupon);
const fromCoupon = (f) => ({
  code: f.code,
  description: f.description,
  description_ar: f.descriptionAr,
  discount_type: f.discountType,
  discount_value: num(f.discountValue),
  scope: f.scope,
  product_ids: f.scope === "products" ? listOrNull(f.productIds) : [],
  category: f.scope === "category" ? f.category : null,
  min_order_amount: num(f.minOrder),
  max_discount: f.discountType === "percentage" && f.maxDiscount !== "" ? num(f.maxDiscount) : null,
  usage_limit: f.usageLimit === "" ? null : Number(f.usageLimit),
  per_customer_limit: f.perCustomerLimit === "" ? null : Number(f.perCustomerLimit),
  is_active: f.isActive !== false,
  ...fromRange(f),
});
export const saveCoupon = (id, f) =>
  api(id ? `/api/admin/coupons/${id}` : "/api/admin/coupons", { method: id ? "PUT" : "POST", body: fromCoupon(f) });
export const deleteCoupon = (id) => api(`/api/admin/coupons/${id}`, { method: "DELETE" });

export const fetchDebtAccounts = async (outstandingOnly) =>
  (await fetchAll(`/api/admin/debts${outstandingOnly ? "?outstanding=true" : ""}`)).map(toDebtAccount);
export const fetchStatement = async (userId) => {
  const res = await api(`/api/admin/customers/${userId}/debt?limit=50`);
  return { account: toDebtAccount(res.account), transactions: (res.transactions?.items ?? []).map(toDebtTransaction) };
};
export const setCreditLimit = (userId, limit) =>
  api(`/api/admin/customers/${userId}/credit-limit`, { method: "PUT", body: { credit_limit: num(limit) } });
export const recordPayment = (userId, amount, note) =>
  api(`/api/admin/customers/${userId}/debt/payments`, { method: "POST", body: { amount: num(amount), note: note || null } });
export const recordAdjustment = (userId, amount, note) =>
  api(`/api/admin/customers/${userId}/debt/adjustments`, { method: "POST", body: { amount: num(amount), note } });

export const fetchAdminWaitingList = async (status) =>
  (await fetchAll(`/api/admin/waiting-list${status ? `?status=${status}` : ""}`)).map(toWaiting);
export const setWaitingStatus = (id, status) =>
  api(`/api/admin/waiting-list/${id}/status`, { method: "PUT", body: { status } });
