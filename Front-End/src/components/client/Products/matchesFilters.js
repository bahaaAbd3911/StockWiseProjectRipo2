import { finalPrice } from "../../../utils/format";

// Applies the shared search filters (search bar, categories, price, rating, stock).
export function matchesFilters(product, filters) {
  const { productName, productCategoty, productPrice, productRate, isStock } = filters;
  const query = (productName || "").trim().toLowerCase();
  const price = finalPrice(product);

  const matchesName =
    !query ||
    product.name.toLowerCase().includes(query) ||
    (product.nameAr ?? "").includes(productName.trim());
  const matchesCategory = !productCategoty || product.category === productCategoty;
  const matchesPrice =
    !Array.isArray(productPrice) ||
    productPrice.length !== 2 ||
    (price >= Number(productPrice[0]) && price <= Number(productPrice[1]));
  const matchesRating = !productRate || Number(product.rating) >= Number(productRate);
  const matchesStock = !isStock || !product.outOfStock;

  return matchesName && matchesCategory && matchesPrice && matchesRating && matchesStock;
}
