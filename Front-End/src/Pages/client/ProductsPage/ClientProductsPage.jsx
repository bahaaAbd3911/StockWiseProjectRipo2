import SearchBar from "../../../components/common/SearchBar/SearchBar";
import UserSettings from "../../../components/common/UserSettings/UserSettings";
import Advertisement from "../../../components/client/Advertisement/Advertisement";
import Products from "../../../components/client/Products/Products";
import SearchFilters from "../../../components/client/SearchFilters/SearchFilters";
import CatalogStatus from "../../../components/client/CatalogStatus/CatalogStatus";
import Footer from "../../../components/common/Footer/Footer";
import img1 from "../../../assets/images/perfume.jpg";

import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import { searchAndResultContainerClass, arbSearchAndResultContainerClass } from "./ClientPage.module.css";

const AD = {
  arbAdTitle: "تجربتك المثالية تبدأ هنا",
  adTitle: "Your Perfect Experience Starts Here",
  arbAdSubTitle: "اكتشف تشكيلتنا المتنوعة،",
  adSubTitle: "Discover our diverse range,",
  arbCom: "واعثر على ما يناسب ذوقك وميزانيتك بكل سهولة.",
  com: "Find what suits your taste and budget with ease.",
  imgSrc: img1,
};

function ClientProductsPage() {
  const { products } = useCatalog();
  const { isAr } = useLang();

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
