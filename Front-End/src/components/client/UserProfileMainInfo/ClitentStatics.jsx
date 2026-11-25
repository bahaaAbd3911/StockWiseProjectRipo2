import { motion } from "motion/react";

import { ClitentStaticsContainerClass, staticClass } from "./UserProfileMainInfo.module.css";
import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";

// Counters from the user's real cart, favorites and orders.
function ClitentStatics() {
  const { orders, favorites, cartCount } = useShop();
  const { tr } = useLang();

  const stats = [
    { icon: "fa-solid fa-bag-shopping", color: "#007BFF", count: orders.length, label: tr("Total Orders", "عدد الطلبات") },
    { icon: "fa-solid fa-heart", color: "#FF4D67", count: favorites.length, label: tr("Favorite Products", "المنتجات المفضلة") },
    { icon: "fa-solid fa-cart-shopping", color: "#16A34A", count: cartCount, label: tr("Items In Cart", "عناصر في السلة") },
  ];

  return (
    <motion.div
      className={ClitentStaticsContainerClass}
      initial={{ opacity: 0, y: -40 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
    >
      {stats.map((item) => (
        <div
          className={staticClass}
          style={{ background: `color-mix(in srgb, ${item.color} 12%, var(--color-surface))`, color: item.color }}
          key={item.label}
        >
          <div style={{ color: item.color }}>
            <i className={item.icon}></i>
          </div>
          <div>
            <div>{item.count}</div>
            <div>{item.label}</div>
          </div>
        </div>
      ))}
    </motion.div>
  );
}

export default ClitentStatics;
