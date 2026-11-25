import {
  faGauge,
  faBox,
  faFolder,
  faCartShopping,
  faUsers,
  faGear,
} from "@fortawesome/free-solid-svg-icons";

export const translations = {
  en: {
    dashboard: "Dashboard",
    products: "Products",
    categories: "Categories",
    orders: "Orders",
    users: "Users",
    website: "WEBSITE",
    settings: "Store Settings",
    logout: "Logout",
  },
  arb: {
    dashboard: "لوحة التحكم",
    products: "المنتجات",
    categories: "التصنيفات",
    orders: "الطلبات",
    users: "المستخدمون",
    website: "إعدادات الموقع",
    settings: "إعدادات المتجر",
    logout: "تسجيل الخروج",
  },
};

// Each entry maps to a nested route declared in App.jsx.
export const mainPages = [
  { key: "dashboard", path: "/admin", icon: faGauge, end: true },
  { key: "products", path: "/admin/products", icon: faBox },
  { key: "categories", path: "/admin/categories", icon: faFolder },
  { key: "orders", path: "/admin/orders", icon: faCartShopping },
  { key: "users", path: "/admin/users", icon: faUsers },
];

export const websitePages = [
  { key: "settings", path: "/admin/settings", icon: faGear },
];
