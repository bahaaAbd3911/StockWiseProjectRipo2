import { isHexColor } from "../../../utils/validators";

export const EMPTY_PRODUCT = {
  name: "",
  nameAr: "",
  image: "",
  images: [],
  rating: 0,
  price: "",
  discountValue: 0,
  description: "",
  arabicDescription: "",
  category: "",
  count: 0,
  outOfStock: false,
  availableColors: [],
};

export const EMPTY_COLOR = { name: "", nameAr: "", hex: "#22C55E" };

export function validateProduct(form, tr) {
  const errors = {};
  const price = Number(form.price);
  const discount = Number(form.discountValue);
  const rating = Number(form.rating);
  const count = Number(form.count);

  if (!form.name.trim()) errors.name = tr("Name is required", "الاسم مطلوب");
  if (!form.nameAr.trim()) errors.nameAr = tr("Arabic name is required", "الاسم العربي مطلوب");
  if (!form.image.trim()) errors.image = tr("Main image URL is required", "رابط الصورة الرئيسية مطلوب");
  if (!form.category) errors.category = tr("Choose a category", "اختر تصنيفًا");
  if (!(price > 0)) errors.price = tr("Price must be greater than 0", "يجب أن يكون السعر أكبر من 0");
  if (discount < 0 || discount > 100) errors.discountValue = tr("Discount must be 0–100", "الخصم بين 0 و 100");
  if (rating < 0 || rating > 5) errors.rating = tr("Rating must be 0–5", "التقييم بين 0 و 5");
  if (!Number.isInteger(count) || count < 0) errors.count = tr("Count must be a whole number ≥ 0", "الكمية عدد صحيح ≥ 0");

  const badColor = form.availableColors.some((c) => !c.name.trim() || !isHexColor(c.hex));
  if (badColor) errors.availableColors = tr("Each color needs a name and a valid hex", "كل لون يحتاج اسمًا ورمز hex صحيحًا");

  return errors;
}

// Converts the form state into the stored product shape.
export function toProduct(form, categories) {
  const category = categories.find((c) => c.name === form.category);
  const count = Number(form.count);

  return {
    ...form,
    name: form.name.trim(),
    nameAr: form.nameAr.trim(),
    image: form.image.trim(),
    images: form.images.map((url) => url.trim()).filter(Boolean),
    rating: Number(form.rating),
    price: Number(form.price),
    discountValue: Number(form.discountValue),
    categoryAr: category?.nameAr ?? "",
    icon: category?.icon ?? "",
    count,
    outOfStock: form.outOfStock || count === 0,
  };
}
