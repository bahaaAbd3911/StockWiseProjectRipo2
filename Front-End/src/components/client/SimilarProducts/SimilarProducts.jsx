import {
  SimilarProductsContainerClass,
  arbSimilarProductsContainerClass,
} from "./SimilarProducts.module.css";

import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import Products from "../Products/Products";

// Other products from the same category.
function SimilarProducts({ product }) {
  const { products } = useCatalog();
  const { tr, isAr } = useLang();

  const similar = products
    .filter((p) => p.category === product.category && p.id !== product.id)
    .slice(0, 8);

  if (similar.length === 0) return null;

  return (
    <div className={`${SimilarProductsContainerClass} ${isAr ? arbSimilarProductsContainerClass : ""}`}>
      <h1>{tr("Similar Products", "منتجات مشابهة")}</h1>
      <Products productsProp={similar} />
    </div>
  );
}

export default SimilarProducts;
