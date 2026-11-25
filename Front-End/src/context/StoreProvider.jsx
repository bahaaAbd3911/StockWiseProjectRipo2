import { useContext, useEffect, useState } from "react";

import { AuthContext } from "./AuthContext";
import { StoreContext } from "./StoreContext";
import { fetchSettings, saveSettings } from "../api/catalog";
import { withDefaults } from "../data/storeSettings";
import { useStoreTheme } from "../hooks/useStoreTheme";

const CACHE_KEY = "storeSettings";

// Last known settings, so the right colors show before the API answers.
const cached = () => {
  try {
    return withDefaults(JSON.parse(localStorage.getItem(CACHE_KEY)));
  } catch {
    return withDefaults();
  }
};

const remember = (settings) => {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify(settings));
  } catch {
    // A large logo can exceed the browser quota; the database copy is what matters.
  }
};

// Store name, logo and theme colors (admin-editable, stored in the database).
export function StoreProvider({ children }) {
  const { dark } = useContext(AuthContext);
  const [storeSettings, setStoreSettings] = useState(cached);

  useStoreTheme(storeSettings, dark);

  useEffect(() => {
    fetchSettings()
      .then((data) => {
        const settings = withDefaults(data);
        setStoreSettings(settings);
        remember(settings);
      })
      .catch(() => {});
  }, []);

  // Throws with the server message on failure.
  const saveStoreSettings = async (settings) => {
    const saved = withDefaults(await saveSettings(settings));
    setStoreSettings(saved);
    remember(saved);
    return saved;
  };

  return (
    <StoreContext.Provider value={{ storeSettings, saveStoreSettings }}>
      {children}
    </StoreContext.Provider>
  );
}
