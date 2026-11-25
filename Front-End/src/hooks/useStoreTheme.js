import { useEffect } from "react";
import { COLOR_TOKENS } from "../data/storeSettings";

// Writes the active mode's colors into the --color-* variables on <html>.
// Every stylesheet reads colors from these variables, so this re-themes the whole app.
export function useStoreTheme(settings, dark) {
  useEffect(() => {
    const colors = settings.colors[dark ? "dark" : "light"];
    const root = document.documentElement;

    COLOR_TOKENS.forEach(({ key, css }) => root.style.setProperty(css, colors[key]));
    root.dataset.theme = dark ? "dark" : "light";
    root.style.colorScheme = dark ? "dark" : "light";

    document.title = settings.storeName;
  }, [settings, dark]);
}
