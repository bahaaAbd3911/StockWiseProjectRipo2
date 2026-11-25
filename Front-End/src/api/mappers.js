// Converts API (snake_case) objects into the shapes the components use, and back.
// Backend shapes: see backend/docs/GUIDE.md, section 12 (API reference).

const num = (value) => Number(value) || 0;

const toColor = (c) => ({ name: c.name, nameAr: c.name_ar || c.name, hex: c.hex_code });

export const toProduct = (p) => {
  const images = p.image_urls?.length ? p.image_urls : p.primary_image_url ? [p.primary_image_url] : [];
  return {
    id: p.id,
    name: p.name,
    nameAr: p.name_ar,
    image: p.primary_image_url ?? images[0] ?? "",
    images,
    rating: num(p.rating_average),
    reviewCount: p.review_count ?? 0,
    price: num(p.price),
    discountValue: num(p.discount_value),
    // What the customer actually pays per unit: the product discount or a live
    // offer, whichever is lower (computed by the backend).
    finalPrice: num(p.final_price ?? p.price),
    offer: p.offer ?? null,
    description: p.description ?? "",
    arabicDescription: p.description_ar ?? "",
    category: p.category ?? "",
    categoryAr: p.category_ar ?? "",
    availableColors: (p.colors ?? []).map(toColor),
    count: p.stock,
    outOfStock: p.in_stock === false || p.stock <= 0,
    isActive: p.is_active !== false,
    createdAt: p.created_at,
  };
};

// Product form -> body of POST/PUT /api/products. Images and colors are saved
// together with the product (the first image becomes the main one). The
// rating is not sent: it comes from customer reviews.
export const fromProduct = (p) => ({
  name: p.name,
  name_ar: p.nameAr,
  description: p.description ?? "",
  description_ar: p.arabicDescription ?? "",
  category: p.category ?? "",
  category_ar: p.categoryAr ?? "",
  price: num(p.price),
  discount_value: num(p.discountValue),
  stock: num(p.count),
  images: [p.image, ...(p.images ?? [])].map((url) => (url ?? "").trim()).filter(Boolean),
  colors: (p.availableColors ?? []).map((c) => ({ name: c.name, name_ar: c.nameAr || c.name, hex_code: c.hex })),
});

export const toCategory = (c) => ({
  id: c.id,
  name: c.name,
  nameAr: c.name_ar,
  icon: c.icon,
  isActive: c.is_active,
  productsCount: c.product_count,
  isSelected: false,
  createdAt: c.created_at,
});

export const fromCategory = (c) => ({
  name: c.name,
  name_ar: c.nameAr,
  icon: c.icon,
  is_active: c.isActive !== false,
});

// Works for the profile (GET /api/profile) and for the admin user list, which
// adds is_active, orders_count and total_spent.
export const toUser = (u) => ({
  userID: u.id,
  firstName: u.first_name,
  lastName: u.last_name,
  userFullName: `${u.first_name} ${u.last_name}`.trim(),
  userEmail: u.email,
  phone: u.phone ?? "",
  image: u.profile_image_url ?? "",
  role: u.role,
  status: u.is_active === false ? "inactive" : "active",
  emailVerified: u.email_verified !== false,
  joinedAt: u.created_at,
  ordersCount: u.orders_count ?? 0,
  totalSpent: num(u.total_spent),
});

// "Sara Al Mustafa" -> first "Sara", last "Al Mustafa"
export const splitName = (fullName) => {
  const [first = "", ...rest] = fullName.trim().split(/\s+/);
  return { first_name: first, last_name: rest.join(" ") };
};

export const toOrderItem = (i) => ({
  id: i.id,
  productID: i.product_id,
  name: i.product_name,
  nameAr: i.product_name_ar || i.product_name,
  image: i.product_image_url ?? "",
  unitPrice: num(i.unit_price),
  quantity: i.quantity,
  color: i.selected_color,
});

// Bundles in an order are shown as one line each.
const toOrderBundle = (b) => ({
  id: `bundle-${b.id}`,
  productID: null,
  name: b.bundle_name,
  nameAr: b.bundle_name_ar || b.bundle_name,
  image: "",
  unitPrice: num(b.unit_price),
  quantity: b.quantity,
  color: null,
});

export const toOrder = (o) => {
  const customer = o.customer ? `${o.customer.first_name} ${o.customer.last_name}`.trim() : "";
  return {
    id: o.id,
    userID: o.user_id,
    customerName: customer || o.recipient_name,
    customerEmail: o.customer?.email ?? "",
    recipientName: o.recipient_name,
    customerPhone: o.phone,
    address: [o.city, o.area, o.street, o.building_number].filter(Boolean).join(", "),
    notes: o.additional_instructions ?? "",
    paymentMethod: o.payment_method,
    paymentStatus: o.payment_status,
    status: o.status,
    subtotal: num(o.subtotal_amount ?? o.total_amount),
    discount: num(o.discount_amount),
    couponCode: o.coupon_code ?? null,
    total: num(o.total_amount),
    createdAt: o.created_at,
    items: [...(o.items ?? []).map(toOrderItem), ...(o.bundles ?? []).map(toOrderBundle)],
  };
};

// Cart lines come with the price the customer pays (unit_price: product
// discount or live offer) and whether they can be bought right now.
export const toCartItem = (i) => ({
  id: i.id,
  productID: i.product_id,
  name: i.product_name,
  nameAr: i.product_name_ar,
  image: i.primary_image_url ?? "",
  price: num(i.original_price),
  discountValue: num(i.discount_value),
  unitPrice: num(i.unit_price),
  lineTotal: num(i.line_total),
  quantity: i.quantity,
  stock: i.stock,
  color: i.selected_color,
  isAvailable: i.is_available !== false,
  unavailableReason: i.unavailable_reason ?? "",
});

// A favorite embeds the whole product.
export const toFavoriteProduct = (f) => toProduct(f.product);

export const toReview = (r) => ({
  id: r.id,
  productID: r.product_id,
  commentTxt: r.comment,
  commentTxtAr: r.comment_ar || r.comment,
  commentWriterID: r.user_id,
  commentWriterName: r.reviewer_name,
  commentWriterImage: "",
  rating: r.rating,
  createdAt: r.created_at,
  verifiedPurchase: r.verified_purchase,
  helpfulCount: r.helpful_count,
  isEdited: r.is_edited,
});
