import { useCallback, useContext, useEffect, useState } from "react";

import { AuthContext } from "./AuthContext";
import { CatalogContext } from "./CatalogContext";
import { fetchCategories, fetchProducts } from "../api/catalog";

// Admins also get switched-off categories, so they can switch them back on.
const loadCatalog = (isAdmin) => Promise.all([fetchProducts(), fetchCategories(isAdmin)]);

export function CatalogProvider({ children }) {
  const { isAdmin } = useContext(AuthContext);
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [status, setStatus] = useState("loading"); // loading | ready | error
  const [error, setError] = useState("");

  const apply = useCallback(([productList, categoryList]) => {
    setProducts(productList);
    setCategories(categoryList);
    setStatus("ready");
  }, []);

  const fail = useCallback((err) => {
    setError(err.message);
    setStatus("error");
  }, []);

  // Re-fetch after admin changes, orders (stock), reviews (ratings)…
  const reload = useCallback(() => loadCatalog(isAdmin).then(apply, fail), [isAdmin, apply, fail]);

  useEffect(() => {
    loadCatalog(isAdmin).then(apply, fail);
  }, [isAdmin, apply, fail]);

  return (
    <CatalogContext.Provider
      value={{ products, setProducts, categories, setCategories, status, error, reload }}
    >
      {children}
    </CatalogContext.Provider>
  );
}
