import { useCallback, useEffect, useState } from "react";

import { CatalogContext } from "./CatalogContext";
import { fetchCategories, fetchProducts } from "../api/catalog";

const loadCatalog = () => Promise.all([fetchProducts(), fetchCategories()]);

export function CatalogProvider({ children }) {
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
  const reload = useCallback(() => loadCatalog().then(apply, fail), [apply, fail]);

  useEffect(() => {
    loadCatalog().then(apply, fail);
  }, [apply, fail]);

  return (
    <CatalogContext.Provider
      value={{ products, setProducts, categories, setCategories, status, error, reload }}
    >
      {children}
    </CatalogContext.Provider>
  );
}
