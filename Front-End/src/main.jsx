import { Suspense, lazy } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.jsx";
import { BrowserRouter } from "react-router";
import { AuthProvider } from "./context/AuthProvider.jsx";
import { StoreProvider } from "./context/StoreProvider.jsx";
import { CatalogProvider } from "./context/CatalogProvider.jsx";
import { ShopProvider } from "./context/ShopProvider.jsx";
import { isPlatformHost } from "./api/platform";
// The platform console is only used on the platform admin host.
// eslint-disable-next-line react-refresh/only-export-components -- entry file, never hot-reloaded as a module
const PlatformApp = lazy(() => import("./Pages/platform/PlatformApp.jsx"));

// One build serves every store: the backend knows the store from the host
// name, and the platform admin host (admin.<domain>) shows the platform
// section instead of a shop.
const root = createRoot(document.getElementById("root"));

if (isPlatformHost()) {
  root.render(
    <BrowserRouter>
      <AuthProvider>
        <Suspense fallback={null}>
          <PlatformApp />
        </Suspense>
      </AuthProvider>
    </BrowserRouter>,
  );
} else {
  root.render(
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
}
