// Import Components
import Categories from "./Categories";
import PriceRangeFilter from "./PriceRangeFilter";
import RatingFilter from "./RatingFilter";
// Import Style
import {
  SearchFiltersContainerClass,
  filtersContainerClass,
  checkBoxClass,
  inStockContainerClass,
  filterContainerClass,
  darkSearchFiltersContainerClass,
  arbSearchFiltersContainerClass,
  notCheckedCheckBoxClass,
  filtersToggleClass,
  filtersOpenClass,
} from "./SearchFilters.module.css";

// Import Hooks
import { useState, useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";
import { useCatalog } from "../../../hooks/useCatalog";

// Import Motion
import { motion } from "motion/react";
function SearchFilters() {
  const { dark, lang, setSearchFilters, searchFilters } =
    useContext(AuthContext);
  const { products } = useCatalog();
  const [open, setOpen] = useState(false);
  // Slider goes up to the most expensive product (rounded up to the next 50).
  const maxPrice = Math.max(50, Math.ceil(Math.max(0, ...products.map((p) => p.price)) / 50) * 50);

  return (
    <>
      <motion.div
        className={`${SearchFiltersContainerClass} ${dark && darkSearchFiltersContainerClass} ${lang === "arb" && arbSearchFiltersContainerClass}`}
        initial={{ opacity: 0, x: -140 }}
        whileInView={{ opacity: 1, x: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.4 }}
      >
        <button
          type="button"
          className={filtersToggleClass}
          aria-expanded={open}
          onClick={() => setOpen((v) => !v)}
        >
          <i className="fa-solid fa-sliders"></i>
          {lang === "arb" ? "خيارات التصفية" : "Filters"}
          <i className={`fa-solid ${open ? "fa-chevron-up" : "fa-chevron-down"}`}></i>
        </button>
        <div className={`${filtersContainerClass} ${open ? filtersOpenClass : ""}`}>
          <Categories />
          <PriceRangeFilter key={maxPrice} min={0} max={maxPrice} />
          <RatingFilter />
          <div className={filterContainerClass}>
            <h4>{lang === "arb" ? "التوفر" : "In Stock"}</h4>
            <div className={inStockContainerClass}>
              <div
                className={`${searchFilters.isStock ? checkBoxClass : notCheckedCheckBoxClass}`}
                onClick={() => {
                  setSearchFilters((prev) => ({
                    ...prev,
                    isStock: !prev.isStock,
                  }));
                }}
              >
                {searchFilters.isStock && <i className="fa-solid fa-check"></i>}
              </div>
              <div>
                {lang === "arb"
                  ? "عرض المنتجات المتوفرة فقط"
                  : "Only Show In Stock"}
              </div>
            </div>
          </div>
        </div>
      </motion.div>
    </>
  );
}
export default SearchFilters;
