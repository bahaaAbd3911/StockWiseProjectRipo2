import {
  faGauge,
  faBox,
  faFolder,
  faCartShopping,
  faUsers,
  faGear,
  faGlobe,
  faTicket,
  faTags,
  faBoxesStacked,
  faWallet,
  faBell,
  faEnvelope,
} from "@fortawesome/free-solid-svg-icons";

export const translations = {
  en: {
    dashboard: "Dashboard",
    products: "Products",
    categories: "Categories",
    orders: "Orders",
    users: "Users",
    coupons: "Coupons",
    offers: "Offers",
    bundles: "Bundles",
    debts: "Customer Debts",
    waiting: "Waiting List",
    emails: "Emails",
    website: "WEBSITE",
    settings: "Store Settings",
    domains: "Domains",
    logout: "Logout",
  },
  arb: {
    dashboard: "لوحة التحكم",
    products: "المنتجات",
    categories: "التصنيفات",
    orders: "الطلبات",
    users: "المستخدمون",
    coupons: "القسائم",
    offers: "العروض",
    bundles: "الباقات",
    debts: "ديون العملاء",
    waiting: "قائمة الانتظار",
    emails: "رسائل البريد",
    website: "إعدادات الموقع",
    settings: "إعدادات المتجر",
    domains: "النطاقات",
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
  { key: "coupons", path: "/admin/coupons", icon: faTicket },
  { key: "offers", path: "/admin/offers", icon: faTags },
  { key: "bundles", path: "/admin/bundles", icon: faBoxesStacked },
  { key: "debts", path: "/admin/debts", icon: faWallet },
  { key: "waiting", path: "/admin/waiting-list", icon: faBell },
  { key: "emails", path: "/admin/emails", icon: faEnvelope },
];

export const websitePages = [
  { key: "settings", path: "/admin/settings", icon: faGear },
  { key: "domains", path: "/admin/domains", icon: faGlobe },
];
