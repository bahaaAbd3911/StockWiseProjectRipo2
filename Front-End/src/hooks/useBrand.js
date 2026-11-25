import { useContext } from "react";
import { StoreContext } from "../context/StoreContext";

// Splits "StockWise" into ["Stock", "Wise"] so the second part can be highlighted.
export function splitBrandName(name) {
  const match = name.match(/^(.+?)([A-Z\s].*)$/);
  return match ? [match[1], match[2]] : [name, ""];
}

export function useBrand() {
  const { storeSettings } = useContext(StoreContext);
  const [first, rest] = splitBrandName(storeSettings.storeName);

  return { name: storeSettings.storeName, logo: storeSettings.logo, first, rest };
}
