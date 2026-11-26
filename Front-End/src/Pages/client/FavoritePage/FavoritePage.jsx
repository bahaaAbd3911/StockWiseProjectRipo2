import { useContext } from "react";

import { AuthContext } from "../../../context/AuthContext";
import { useShop } from "../../../hooks/useShop";
import { useCatalog } from "../../../hooks/useCatalog";
import Header from "../../../components/client/Header/Header";
import Products from "../../../components/client/Products/Products";
import EmptyMessage from "../../../components/client/EmptyMessage/EmptyMessage";
import SignInRequired from "../../../components/client/SignInRequired/SignInRequired";
import favoriteImg from "../../../assets/images/Favorites/favoriteLogo.webp";

// Shown inside the profile page ("My Favorites" tab) and at /client/favoriteList.
function FavoritePage() {
  const { isSign } = useContext(AuthContext);
  const { favorites } = useShop();
  const { products } = useCatalog();

  if (!isSign) return <SignInRequired />;

  // Use the full catalogue entry when available (colors, category…) for the cards.
  const items = favorites.map((f) => products.find((p) => p.id === f.id) ?? f);

  return (
    <div style={{ width: "min(1200px, 94%)", margin: "0 auto" }}>
      <Header
        icon={"fa-solid fa-heart"}
        title={"My Favorites"}
        arbTitle={"قائمة المفضلة"}
        subTitle={"Your saved items are here. Keep track of the products you love."}
        arbSubTitle={"العناصر المحفوظة موجودة هنا. تابع المنتجات التي تحبها."}
        img={favoriteImg}
      />
      <Products
        productsProp={items}
        emptyState={
          <EmptyMessage
            img={favoriteImg}
            title={"No favorites yet"}
            arbTitle={"لا توجد منتجات مفضلة بعد"}
            subTitle={"Tap the heart on any product to save it here."}
            arbSubTitle={"اضغط على القلب في أي منتج لحفظه هنا."}
          />
        }
      />
    </div>
  );
}

export default FavoritePage;
