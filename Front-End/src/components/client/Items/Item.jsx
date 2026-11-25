import { useState } from "react";
import { useNavigate } from "react-router";
import { motion } from "motion/react";

import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import FavoriteButton from "./FavoriteButton";
import PriceTag from "./PriceTag";
import styles from "./Item.module.css";

// Product card used on every product grid in the store.
function Item({ product }) {
  const navigate = useNavigate();
  const { tr, isAr } = useLang();
  const { addToCart, cartItemFor } = useShop();
  const [adding, setAdding] = useState(false);

  const inCart = Boolean(cartItemFor(product.id));
  const openDetails = () => navigate(`/client/ProductDetails/${product.id}`);

  const add = async () => {
    setAdding(true);
    await addToCart(product);
    setAdding(false);
  };

  return (
    <motion.article
      className={styles.itemContainerClass}
      whileHover={{ y: -6 }}
      transition={{ duration: 0.25 }}
    >
      <div className={styles.itemImgClass} onClick={openDetails}>
        <img src={product.image} alt={isAr ? product.nameAr : product.name} loading="lazy" />
        {product.discountValue > 0 && <span className={styles.badgeClass}>-{product.discountValue}%</span>}
        <FavoriteButton product={product} />
      </div>

      <div className={styles.itemContnentContainerClass}>
        <div className={styles.itemNameClass} onClick={openDetails}>
          {isAr ? product.nameAr : product.name}
        </div>

        <div className={styles.itemRatingClass}>
          <i className="fa-solid fa-star"></i>
          <span>{Number(product.rating || 0).toFixed(1)}</span>
        </div>

        <div className={styles.priceAndStockStatusContainerClass}>
          <PriceTag product={product} />
          {product.outOfStock && <span className={styles.outOfStockClass}>{tr("Out of stock", "نفد من المخزون")}</span>}
        </div>

        <button
          onClick={add}
          disabled={product.outOfStock || adding}
          className={inCart ? styles.inCartClass : ""}
        >
          <i className={`fa-solid ${adding ? "fa-spinner fa-spin" : inCart ? "fa-check" : "fa-cart-plus"}`}></i>
          {inCart ? tr("In cart · add one more", "في السلة · أضف واحدًا") : tr("Add To Cart", "أضف للسلة")}
        </button>
      </div>
    </motion.article>
  );
}

export default Item;
