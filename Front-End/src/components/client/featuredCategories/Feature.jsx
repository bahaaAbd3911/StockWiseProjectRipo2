import { useContext } from "react";
import { useNavigate } from "react-router";
import { motion } from "motion/react";

import {
  featureContainerClass,
  categoryIconClass,
  categoryNameClass,
  categoryCountClass,
} from "./featuredCategories.module.css";
import { AuthContext } from "../../../context/AuthContext";
import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";

// Clicking a category opens the products page filtered to it.
function Feature({ category }) {
  const { setSearchFilters } = useContext(AuthContext);
  const { setCategories } = useCatalog();
  const { tr, isAr } = useLang();
  const navigate = useNavigate();

  const open = () => {
    setSearchFilters((prev) => ({ ...prev, productCategoty: category.name }));
    setCategories((prev) => prev.map((c) => ({ ...c, isSelected: c.id === category.id })));
    navigate("/client/products");
  };

  return (
    <motion.button
      type="button"
      className={featureContainerClass}
      whileHover={{ y: -8 }}
      transition={{ duration: 0.25 }}
      onClick={open}
    >
      <div className={`${category.icon} ${categoryIconClass}`}></div>
      <div>
        <div className={categoryNameClass}>{isAr ? category.nameAr : category.name}</div>
        <div className={categoryCountClass}>
          {category.productsCount} {tr("Products", "منتج")}
        </div>
      </div>
    </motion.button>
  );
}

export default Feature;
