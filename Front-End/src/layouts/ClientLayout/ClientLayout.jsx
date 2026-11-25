// Import Style
import {
  ClientPageContainerClass,
  ClientPageContainerDarkModeClass,
} from "../../pages/client/ProductsPage/ClientPage.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../context/AuthContext";

// Import Router
import { Routes, Route, Navigate } from "react-router";

// Import Components
import ClientHomePage from "../../pages/client/HomePage/ClientHomePage";
import ClientProductsPage from "../../pages/client/ProductsPage/ClientProductsPage";
import ShoppingCartPage from "../../pages/client/ShoppingCartPage/ShoppingCartPage";
import ProductDetailsPage from "../../components/client/ProductDetailsPage/ProductDetailsPage";
import PurchaseConfirmation from "../../pages/client/PurchaseConfirmation/PurchaseConfirmation";
import FavoritePage from "../../pages/client/FavoritePage/FavoritePage";
import ClientProfilePage from "../../pages/client/ProfilePage/ClientProfilePage";

function ClientLayout() {
  const { dark } = useContext(AuthContext);

  return (
    <div
      className={`${ClientPageContainerClass} ${
        dark ? ClientPageContainerDarkModeClass : ""
      }`}
    >
      <Routes>
        <Route path="home" element={<ClientHomePage />} />
        <Route path="products" element={<ClientProductsPage />} />
        <Route path="favoriteList" element={<FavoritePage />} />
        <Route path="shoppingList" element={<ShoppingCartPage />} />
        <Route
          path="ProductDetails/:productID"
          element={<ProductDetailsPage />}
        />
        <Route
          path="PurchaseConfirmation"
          element={<PurchaseConfirmation />}
        />
        <Route path="profile" element={<ClientProfilePage />} />

        <Route path="*" element={<Navigate to="/client/home" replace />} />
      </Routes>
    </div>
  );
}

export default ClientLayout;
