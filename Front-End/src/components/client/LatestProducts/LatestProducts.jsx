import { useState } from "react";
import { motion } from "motion/react";

import {
  LatestProductsContainerClass,
  filterContainerClass,
  activeClass,
  arbFilterContainerClass,
  hoveredClass,
} from "./LatestProducts.module.css";
import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import ComponentHeader from "../ComponentHeader/ComponentHeader";
import Products from "../Products/Products";

const SHOWN = 8;

// Newest products, filterable by the store's biggest categories.
function LatestProducts() {
  const { products, categories } = useCatalog();
  const { tr, isAr } = useLang();
  const [active, setActive] = useState("All");

  const topCategories = categories
    .filter((c) => c.isActive !== false && c.productsCount > 0)
    .sort((a, b) => b.productsCount - a.productsCount)
    .slice(0, 5);

  const latest = [...products]
    .sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt) || b.id - a.id)
    .filter((p) => active === "All" || p.category === active)
    .slice(0, SHOWN);

  const filters = [{ name: "All", nameAr: "الكل" }, ...topCategories];

  return (
    <motion.div
      className={LatestProductsContainerClass}
      initial={{ opacity: 0, y: 40 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.1 }}
      transition={{ duration: 0.4 }}
    >
      <ComponentHeader title="Latest Products" arbicTitle="آخر المنتجات" />

      <div className={`${filterContainerClass} ${isAr ? arbFilterContainerClass : ""}`} role="tablist">
        {filters.map((filter) => (
          <button
            type="button"
            role="tab"
            aria-selected={active === filter.name}
            key={filter.name}
            className={active === filter.name ? activeClass : hoveredClass}
            onClick={() => setActive(filter.name)}
          >
            {isAr ? filter.nameAr : filter.name}
          </button>
        ))}
      </div>

      <Products productsProp={latest} />
      {latest.length === 0 && <p style={{ textAlign: "center" }}>{tr("No products yet.", "لا توجد منتجات بعد.")}</p>}
    </motion.div>
  );
}

export default LatestProducts;
