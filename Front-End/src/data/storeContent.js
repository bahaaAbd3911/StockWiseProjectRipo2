// The content a store shows when its admin has not customised it (Store
// Settings -> Page content). Every field an admin fills in replaces the value
// here, for that store only; empty fields keep these. The texts are the
// template's original ones, so a new store looks complete from the start.
import heroImage from "../assets/images/istockphoto-1170073824-612x612.jpg";
import productsBannerImage from "../assets/images/perfume.jpg";

export const DEFAULT_HERO = {
  title: "Everything You Want On One Place",
  titleAr: "كل ما تريده ي مكان واحد",
  subtitle: "Discover Amazing Products, great price and",
  subtitleAr: "اكتشف منتجات مذهلة، وأسعار رائعة،",
  highlight: "a seamless shopping experience",
  highlightAr: "وتجربة تسوق سلسة.",
  buttonText: "Shop Now",
  buttonTextAr: "تسوق الآن",
  image: heroImage,
};

export const DEFAULT_PRODUCTS_BANNER = {
  title: "Your Perfect Experience Starts Here",
  titleAr: "تجربتك المثالية تبدأ هنا",
  subtitle: "Discover our diverse range,",
  subtitleAr: "اكتشف تشكيلتنا المتنوعة،",
  highlight: "Find what suits your taste and budget with ease.",
  highlightAr: "واعثر على ما يناسب ذوقك وميزانيتك بكل سهولة.",
  buttonText: "Shop Now",
  buttonTextAr: "تسوق الآن",
  image: productsBannerImage,
};

export const DEFAULT_FEATURES = [
  { icon: "fa-solid fa-truck", name: "Free Shipping", nameAr: "توصيل مجاني", comment: "On Orders Over $$0", commentAr: "لكل الطلبات" },
  { icon: "fa-regular fa-eye-slash", name: "Secure Payment", nameAr: "دفع آمن", comment: "Cash On", commentAr: "عند الاستلام" },
  { icon: "fa-solid fa-clock-rotate-left", name: "Easy Return", nameAr: "سهولة إرجاع الطلبات", comment: "Withen 7 Days", commentAr: "خلال سبع أيام" },
  { icon: "fa-solid fa-business-time", name: "24/7 Support", nameAr: "متاحون على مدار الساعة", comment: "We Are Here To Help", commentAr: "جاهون للمساعدة" },
];

export const DEFAULT_FOOTER = {
  tagline: "",
  taglineAr: "",
  copyright: "2005 StockWise. All rights reserved",
  copyrightAr: "2005 ستوك وايز. جميع الحقوق محفوظة",
  social: {},
};

// Social networks a footer can link to: key -> Font Awesome icon.
export const SOCIAL_NETWORKS = {
  facebook: "fa-brands fa-facebook",
  instagram: "fa-brands fa-instagram",
  x: "fa-brands fa-x-twitter",
  youtube: "fa-brands fa-youtube",
  tiktok: "fa-brands fa-tiktok",
  whatsapp: "fa-brands fa-whatsapp",
  snapchat: "fa-brands fa-snapchat",
  linkedin: "fa-brands fa-linkedin",
};

// Fills a banner's empty fields from its defaults.
export const withBannerDefaults = (banner, defaults) => {
  const out = { ...defaults };
  Object.entries(banner ?? {}).forEach(([key, value]) => {
    if (typeof value === "string" && value.trim()) out[key] = value;
  });
  return out;
};
