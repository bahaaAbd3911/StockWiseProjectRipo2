// Import Components
import Category from "./Category";

// Import Style
import {
  categoriesContainerClass,
  filterContainerClass,
} from "./SearchFilters.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";
import { useCatalog } from "../../../hooks/useCatalog";

function Categories() {
  const { lang } = useContext(AuthContext);
  const { categories, setCategories } = useCatalog();
  const visible = categories.filter((c) => c.isActive !== false);

  return (
    <>
      <div className={`${filterContainerClass} `}>
        <h4>{lang === "arb" ? "الفئات" : "Categories"}</h4>
        <div className={categoriesContainerClass}>
          {visible.map((categoryData, index) => (
            <Category
              data={categoryData}
              key={index}
              setCategories={setCategories}
            />
          ))}
        </div>
      </div>
    </>
  );
}

export default Categories;
