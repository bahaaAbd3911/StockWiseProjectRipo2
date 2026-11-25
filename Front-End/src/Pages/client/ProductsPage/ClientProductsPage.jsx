import SearchBar from "../../../components/common/SearchBar/SearchBar";
import UserSettings from "../../../components/common/UserSettings/UserSettings";
import Advertisement from "../../../components/client/Advertisement/Advertisement";
import Products from "../../../components/client/Products/Products";
import SearchFilters from "../../../components/client/SearchFilters/SearchFilters";
import CatalogStatus from "../../../components/client/CatalogStatus/CatalogStatus";
import Footer from "../../../components/common/Footer/Footer";
import { toAdData, useStoreContent } from "../../../hooks/useStoreContent";

import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import { searchAndResultContainerClass, arbSearchAndResultContainerClass } from "./ClientPage.module.css";

function ClientProductsPage() {
  const { products } = useCatalog();
  const { isAr } = useLang();
  // The store's own banner (Store Settings), or the defaults.
  const AD = toAdData(useStoreContent().productsBanner);

  return (
    <div style={{ overflowX: "hidden" }}>
      <UserSettings />
      <SearchBar value={1} />
      <Advertisement data={AD} />
      <div className={`${searchAndResultContainerClass} ${isAr ? arbSearchAndResultContainerClass : ""}`}>
        <SearchFilters />
        <CatalogStatus>
          <Products productsProp={products} applyFilters />
        </CatalogStatus>
      </div>
      <Footer />
    </div>
  );
}

export default ClientProductsPage;
