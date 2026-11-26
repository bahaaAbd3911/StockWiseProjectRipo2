import { useState } from "react";

import { useShop } from "../../../hooks/useShop";
import { useToast } from "../../../hooks/useToast";
import { joinWaitingList } from "../../../api/sales";
import { useLang } from "../../../hooks/useLang";
import { btnContainerClass, favoriteOnClass } from "./ProductDetails.module.css";

function ProductButtons({ productData, quantity, color }) {
  const { tr } = useLang();
  const { addToCart, toggleFavorite, isFavorite, requireSignIn } = useShop();
  const toast = useToast();
  const [adding, setAdding] = useState(false);
  const favorite = isFavorite(productData.id);

  const add = async () => {
    setAdding(true);
    await addToCart(productData, quantity, color);
    setAdding(false);
  };

  // Out of stock: ask to be emailed when it comes back.
  const notify = async () => {
    if (!requireSignIn()) return;
    setAdding(true);
    try {
      await joinWaitingList(productData.id, quantity, color);
      toast.success("We will email you when it is back in stock", "سنرسل لك بريداً عند توفره");
    } catch (err) {
      toast.error(err.message, err.message);
    }
    setAdding(false);
  };

  return (
    <div className={btnContainerClass}>
      {productData.outOfStock ? (
        <button type="button" onClick={notify} disabled={adding}>
          <i className={`fa-solid ${adding ? "fa-spinner fa-spin" : "fa-bell"}`}></i>
          {tr("Out of stock · Notify me", "نفد من المخزون · أعلمني")}
        </button>
      ) : (
        <button type="button" onClick={add} disabled={adding}>
          <i className={`fa-solid ${adding ? "fa-spinner fa-spin" : "fa-cart-shopping"}`}></i>
          {tr("Add To Cart", "أضف الى سلة التسوق")}
        </button>
      )}
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
