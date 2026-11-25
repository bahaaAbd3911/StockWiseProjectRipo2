import { useState } from "react";

import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import { btnContainerClass, favoriteOnClass } from "./ProductDetails.module.css";

function ProductButtons({ productData, quantity, color }) {
  const { tr } = useLang();
  const { addToCart, toggleFavorite, isFavorite } = useShop();
  const [adding, setAdding] = useState(false);
  const favorite = isFavorite(productData.id);

  const add = async () => {
    setAdding(true);
    await addToCart(productData, quantity, color);
    setAdding(false);
  };

  return (
    <div className={btnContainerClass}>
      <button type="button" onClick={add} disabled={productData.outOfStock || adding}>
        <i className={`fa-solid ${adding ? "fa-spinner fa-spin" : "fa-cart-shopping"}`}></i>
        {productData.outOfStock ? tr("Out of stock", "نفد من المخزون") : tr("Add To Cart", "أضف الى سلة التسوق")}
      </button>
      <button
        type="button"
        className={favorite ? favoriteOnClass : ""}
        aria-pressed={favorite}
        onClick={() => toggleFavorite(productData)}
      >
        <i className={`${favorite ? "fa-solid" : "fa-regular"} fa-heart`}></i>
        {favorite ? tr("In your favorites", "في قائمة المفضلة") : tr("Add To Favorites", "أضف الى المفضلة")}
      </button>
    </div>
  );
}

export default ProductButtons;
