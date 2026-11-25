import { useContext } from "react";
import { motion } from "motion/react";

import { AuthContext } from "../../../context/AuthContext";
import { useLang } from "../../../hooks/useLang";
import Item from "../Items/Item";
import { matchesFilters } from "./matchesFilters";
import { ProductsContainerClass, noProductsMatchingClass } from "./Products.module.css";

// Product grid. `applyFilters` turns on the shared search filters (products page).
function Products({ productsProp = [], applyFilters = false, emptyState = null }) {
  const { searchFilters } = useContext(AuthContext);
  const { tr } = useLang();

  const products = applyFilters ? productsProp.filter((p) => matchesFilters(p, searchFilters)) : productsProp;

  if (productsProp.length === 0 && emptyState) return emptyState;

  return (
    <motion.div
      className={ProductsContainerClass}
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
    >
      {products.length > 0 ? (
        products.map((product) => <Item product={product} key={product.id} />)
      ) : (
        <p className={noProductsMatchingClass}>
          {tr("No products match your filters.", "لا توجد منتجات تطابق خيارات البحث.")}
        </p>
      )}
    </motion.div>
  );
}

export default Products;
