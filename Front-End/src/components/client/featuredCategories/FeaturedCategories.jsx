import { motion } from "motion/react";

import {
  featuredCategoriesContainerClass,
  featuredCategoriesClass,
  arbFeaturedCategoriesClass,
} from "./featuredCategories.module.css";
import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import ComponentHeader from "../ComponentHeader/ComponentHeader";
import Feature from "./Feature";

// The categories with the most products (live counts from the database).
function FeaturedCategories() {
  const { categories } = useCatalog();
  const { isAr } = useLang();

  const featured = categories
    .filter((c) => c.isActive !== false && c.productsCount > 0)
    .sort((a, b) => b.productsCount - a.productsCount)
    .slice(0, 6);

  if (featured.length === 0) return null;

  return (
    <motion.div
      className={featuredCategoriesContainerClass}
      initial={{ opacity: 0, y: 25 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.1 }}
      transition={{ duration: 0.4 }}
    >
      <ComponentHeader title={"Featured Categories"} arbicTitle={"الفئات المميزة"} />
      <div className={`${featuredCategoriesClass} ${isAr ? arbFeaturedCategoriesClass : ""}`}>
        {featured.map((category) => (
          <Feature category={category} key={category.id} />
        ))}
      </div>
    </motion.div>
  );
}

export default FeaturedCategories;
