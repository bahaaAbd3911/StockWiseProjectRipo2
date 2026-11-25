import { useContext } from "react";
import { AuthContext } from "../context/AuthContext";

// Shared language/theme helpers: tr("English", "عربي") returns the active one.
export function useLang() {
  const { lang, dark } = useContext(AuthContext);
  const isAr = lang === "arb";

  return {
    isAr,
    dark,
    dir: isAr ? "rtl" : "ltr",
    tr: (en, ar) => (isAr ? ar : en),
  };
}
