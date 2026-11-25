// Converts API (snake_case) objects into the shapes the components use, and back.

const num = (value) => Number(value) || 0;

export const toProduct = (p) => ({
  id: p.id,
  name: p.name,
  nameAr: p.name_ar,
  image: p.image_url,
  images: p.images?.length ? p.images : p.image_url ? [p.image_url] : [],
  rating: num(p.rating),
  price: num(p.price),
  discountValue: num(p.discount_value),
  description: p.description,
  arabicDescription: p.description_ar,
  category: p.category,
  categoryAr: p.category_ar,
  availableColors: (p.colors ?? []).map((c) => ({ name: c.name, nameAr: c.name_ar, hex: c.hex })),
  count: p.stock,
  outOfStock: p.stock <= 0,
  createdAt: p.created_at,
});

export const fromProduct = (p) => ({
  name: p.name,
  name_ar: p.nameAr,
  description: p.description ?? "",
  description_ar: p.arabicDescription ?? "",
  category: p.category,
  price: num(p.price),
  discount_value: num(p.discountValue),
  stock: num(p.count),
  image_url: p.image,
  images: p.images ?? [],
  rating: num(p.rating),
  colors: (p.availableColors ?? []).map((c) => ({ name: c.name, name_ar: c.nameAr || c.name, hex: c.hex })),
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

export const toUser = (u) => ({
  userID: u.id,
  firstName: u.first_name,
  lastName: u.last_name,
  userFullName: `${u.first_name} ${u.last_name}`.trim(),
  userEmail: u.email,
  phone: u.phone,
  image: u.profile_image_url ?? "",
  role: u.role,
  status: u.is_active === false ? "inactive" : "active",
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
  image: i.product_image,
  unitPrice: num(i.unit_price),
  quantity: i.quantity,
  color: i.selected_color,
});

export const toOrder = (o) => ({
  id: o.id,
  userID: o.user_id,
  customerName: o.customer_name || o.recipient_name,
  customerEmail: o.customer_email ?? "",
  recipientName: o.recipient_name,
  customerPhone: o.phone,
  address: [o.city, o.area, o.street, o.building_number].filter(Boolean).join(", "),
  notes: o.additional_instructions ?? "",
  paymentMethod: o.payment_method,
  paymentStatus: o.payment_status,
  status: o.status,
  total: num(o.total_amount),
  createdAt: o.created_at,
  items: (o.items ?? []).map(toOrderItem),
});

export const toCartItem = (i) => ({
  id: i.id,
  productID: i.product_id,
  name: i.product_name,
  nameAr: i.product_name_ar,
  image: i.image_url,
  price: num(i.price),
  discountValue: num(i.discount_value),
  unitPrice: num(i.price) * (1 - num(i.discount_value) / 100),
  quantity: i.quantity,
  stock: i.stock,
  color: i.selected_color,
});

// Favorites come back in a slim shape; give them the product-card fields.
export const toFavoriteProduct = (f) => ({
  id: f.product_id,
  name: f.product_name,
  nameAr: f.product_name_ar,
  image: f.image_url,
  images: [f.image_url],
  price: num(f.price),
  discountValue: num(f.discount_value),
  rating: num(f.rating),
  count: f.stock,
  outOfStock: f.stock <= 0,
});

export const toReview = (r) => ({
  id: r.id,
  productID: r.product_id,
  commentTxt: r.comment,
  commentTxtAr: r.comment_ar || r.comment,
  commentWriterID: r.user_id,
  commentWriterName: r.user_name,
  commentWriterImage: r.user_image ?? "",
  rating: r.rating,
  createdAt: r.created_at,
  verifiedPurchase: r.verified_purchase,
  helpfulCount: r.helpful_count,
  isEdited: r.is_edited,
});
