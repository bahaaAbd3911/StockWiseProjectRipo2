// Import Style
import {
  ClientPageContainerClass,
  ClientPageContainerDarkModeClass,
} from "../../Pages/client/ProductsPage/ClientPage.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../context/AuthContext";

// Import Router
import { Routes, Route, Navigate } from "react-router";

// Import Components
import ClientHomePage from "../../Pages/client/HomePage/ClientHomePage";
import ClientProductsPage from "../../Pages/client/ProductsPage/ClientProductsPage";
import ShoppingCartPage from "../../Pages/client/ShoppingCartPage/ShoppingCartPage";
import ProductDetailsPage from "../../components/client/ProductDetailsPage/ProductDetailsPage";
import PurchaseConfirmation from "../../Pages/client/PurchaseConfirmation/PurchaseConfirmation";
import FavoritePage from "../../Pages/client/FavoritePage/FavoritePage";
import ClientProfilePage from "../../Pages/client/ProfilePage/ClientProfilePage";
import OffersPage from "../../Pages/client/Sales/OffersPage";
import BundlesPage from "../../Pages/client/Sales/BundlesPage";
import VerifyEmailBanner from "../../components/client/VerifyEmailBanner/VerifyEmailBanner";

function ClientLayout() {
  const { dark } = useContext(AuthContext);

  return (
    <div
      className={`${ClientPageContainerClass} ${
        dark ? ClientPageContainerDarkModeClass : ""
      }`}
    >
      <VerifyEmailBanner />
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
        <Route path="offers" element={<OffersPage />} />
        <Route path="bundles" element={<BundlesPage />} />

        <Route path="*" element={<Navigate to="/client/home" replace />} />
      </Routes>
    </div>
  );
}

export default ClientLayout;
