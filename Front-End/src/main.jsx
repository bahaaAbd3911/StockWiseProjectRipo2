import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.jsx";
import { BrowserRouter } from "react-router";
import { AuthProvider } from "./context/AuthProvider.jsx";
import { StoreProvider } from "./context/StoreProvider.jsx";
import { CatalogProvider } from "./context/CatalogProvider.jsx";
import { ShopProvider } from "./context/ShopProvider.jsx";

createRoot(document.getElementById("root")).render(
  <BrowserRouter>
    <AuthProvider>
      <StoreProvider>
        <CatalogProvider>
          <ShopProvider>
            <App />
          </ShopProvider>
        </CatalogProvider>
      </StoreProvider>
    </AuthProvider>
  </BrowserRouter>,
);
