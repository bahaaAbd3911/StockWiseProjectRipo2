import { motion } from "motion/react";

import { BestSellingProductsContainerClass } from "./BestSellingProducts.module.css";
import { useCatalog } from "../../../hooks/useCatalog";
import ComponentHeader from "../ComponentHeader/ComponentHeader";
import Products from "../Products/Products";

// Highest-rated products in stock (ratings come from customer reviews).
function BestSellingProducts() {
  const { products } = useCatalog();

  const topRated = products
    .filter((p) => !p.outOfStock)
    .sort((a, b) => b.rating - a.rating)
    .slice(0, 4);

  if (topRated.length === 0) return null;

  return (
    <motion.div
      className={BestSellingProductsContainerClass}
      initial={{ opacity: 0, y: 40 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.1 }}
      transition={{ duration: 0.4 }}
    >
      <ComponentHeader title={"Top Rated Products"} arbicTitle={"الأعلى تقييمًا"} />
      <Products productsProp={topRated} />
    </motion.div>
  );
}

export default BestSellingProducts;
