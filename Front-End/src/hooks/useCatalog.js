import { useContext } from "react";
import { CatalogContext } from "../context/CatalogContext";

export const useCatalog = () => useContext(CatalogContext);
