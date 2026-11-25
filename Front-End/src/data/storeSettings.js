// Every color in the app comes from these tokens (see index.css). Admins edit them
// per mode in Store Settings; they are stored in the database.
export const COLOR_TOKENS = [
  { key: "text", css: "--color-text", en: "Font Color", ar: "لون الخط", hintEn: "Main text", hintAr: "النص الرئيسي" },
  { key: "mutedText", css: "--color-text-muted", en: "Secondary Text", ar: "النص الثانوي", hintEn: "Hints, labels, dates", hintAr: "التلميحات والتواريخ" },
  { key: "background", css: "--color-bg", en: "Background Color", ar: "لون الخلفية", hintEn: "Page background", hintAr: "خلفية الصفحة" },
  { key: "surface", css: "--color-surface", en: "Card Color", ar: "لون البطاقات", hintEn: "Cards, panels, forms", hintAr: "البطاقات واللوحات" },
  { key: "border", css: "--color-border", en: "Border Color", ar: "لون الحدود", hintEn: "Borders and separators", hintAr: "الحدود والفواصل" },
  { key: "button", css: "--color-primary", en: "Button Color", ar: "لون الأزرار", hintEn: "Primary buttons & links", hintAr: "الأزرار الرئيسية" },
  { key: "buttonHover", css: "--color-primary-hover", en: "Button Hover", ar: "تمرير الأزرار", hintEn: "Button hover state", hintAr: "عند التمرير" },
  { key: "accent", css: "--color-accent", en: "Accent Color", ar: "اللون المميز", hintEn: "Highlights & badges", hintAr: "العناصر المميزة" },
  { key: "success", css: "--color-success", en: "Success Color", ar: "لون النجاح", hintEn: "In stock, completed", hintAr: "متوفر، مكتمل" },
  { key: "danger", css: "--color-danger", en: "Danger Color", ar: "لون التحذير", hintEn: "Errors, delete, out of stock", hintAr: "الأخطاء والحذف" },
];

export const DEFAULT_STORE_SETTINGS = {
  storeName: "StockWise",
  logo: "",
  colors: {
    light: {
      text: "#17231D",
      mutedText: "#647488",
      background: "#F5F8FA",
      surface: "#FFFFFF",
      border: "#E2EBF0",
      button: "#22C55E",
      buttonHover: "#16A34A",
      accent: "#4ADE80",
      success: "#10B981",
      danger: "#EF4444",
    },
    dark: {
      text: "#E8F5F0",
      mutedText: "#94A69B",
      background: "#0F1A17",
      surface: "#192520",
      border: "#2A3A32",
      button: "#22C55E",
      buttonHover: "#16A34A",
      accent: "#4ADE80",
      success: "#34D399",
      danger: "#F87171",
    },
  },
};

// Fills missing keys (e.g. settings saved before a token existed) from the defaults.
export function withDefaults(settings) {
  const s = settings ?? {};
  return {
    ...DEFAULT_STORE_SETTINGS,
    ...s,
    colors: {
      light: { ...DEFAULT_STORE_SETTINGS.colors.light, ...s.colors?.light },
      dark: { ...DEFAULT_STORE_SETTINGS.colors.dark, ...s.colors?.dark },
    },
  };
}
