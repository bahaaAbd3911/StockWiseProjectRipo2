// Static UI configuration for the client pages (not store data).

export const PROFILE_TABS = [
  { icon: "fa-solid fa-user", label: "My Profile", arbLabel: "الصفحة الشخصية" },
  { icon: "fa-solid fa-box", label: "My Orders", arbLabel: "طلباتي" },
  { icon: "fa-solid fa-cart-shopping", label: "My Cart", arbLabel: "سلة التسوق" },
  { icon: "fa-solid fa-heart", label: "My Favorites", arbLabel: "قائمتي المفضلة" },
  { icon: "fa-solid fa-right-from-bracket", label: "Log Out", arbLabel: "تسجيل الخروج" },
].map((tab, index) => ({ ...tab, index, isActive: index === 0 }));

export const ORDER_STAGES = [
  { num: 1, name: "Shipping & Contact", arbName: "الشحن والتواصل", width: "33%", label: "Shipping" },
  { num: 2, name: "Payment Method", arbName: "طريقة الدفع", width: "66%", label: "Payment" },
  { num: 3, name: "Review And Place Order", arbName: "مراجعة الطلب وتقديمه", width: "100%", label: "Review" },
].map((stage, index) => ({ ...stage, isSelected: index === 0, alert: false }));

// Checkout fields by index: 0 name, 1 phone, 2 email, 3 city, 4 area,
// 5 address details, 6 delivery option, 7 payment method.
export const emptyOrderForm = () =>
  Array.from({ length: 8 }, (_, index) => ({
    data: "",
    flag: false,
    errorMsg: "",
    arbErrorMsg: "",
    index,
  }));

export const EMPTY_SEARCH_FILTERS = {
  productName: "",
  productCategoty: "",
  productPrice: [],
  productRate: "",
  isStock: false,
};
