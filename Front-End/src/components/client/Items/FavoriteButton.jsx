import { useShop } from "../../../hooks/useShop";
import { useLang } from "../../../hooks/useLang";
import styles from "./Item.module.css";

export default function FavoriteButton({ product }) {
  const { isFavorite, toggleFavorite } = useShop();
  const { tr } = useLang();
  const active = isFavorite(product.id);

  return (
    <button
      type="button"
      className={`${styles.favoBtnClass} ${active ? styles.favoriteActiveClass : ""}`}
      aria-pressed={active}
      aria-label={active ? tr("Remove from favorites", "إزالة من المفضلة") : tr("Add to favorites", "أضف إلى المفضلة")}
      onClick={(e) => {
        e.stopPropagation();
        toggleFavorite(product);
      }}
    >
      <i className={`${active ? "fa-solid" : "fa-regular"} fa-heart`}></i>
    </button>
  );
}
