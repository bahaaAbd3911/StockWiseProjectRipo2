// Import Component
import Advertisement from "../../../components/client/Advertisement/Advertisement";
import Features from "../../../components/client/Features/Features";
import FeaturedCategories from "../../../components/client/featuredCategories/FeaturedCategories";
import BestSellingProducts from "../../../components/client/BestSellingProducts/BestSellingProducts";
import LatestProducts from "../../../components/client/LatestProducts/LatestProducts";
import Footer from "../../../components/common/Footer/Footer";
import SearchBar from "../../../components/common/SearchBar/SearchBar";
import UserSettings from "../../../components/common/UserSettings/UserSettings";

import { toAdData, useStoreContent } from "../../../hooks/useStoreContent";

function ClientHomePage() {
  // The store's own banner (Store Settings), or the defaults.
  const data = toAdData(useStoreContent().hero);

  return (
    <>
      <div>
        <UserSettings />
        <SearchBar value={0} />
        <Advertisement data={data} value={0} />
        <Features />
        <FeaturedCategories />
        <BestSellingProducts value={0} />
        <LatestProducts value={0} />
        <Footer />
      </div>
    </>
  );
}

export default ClientHomePage;
