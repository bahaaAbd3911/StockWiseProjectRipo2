// Import Component
import Advertisement from "../../../components/client/Advertisement/Advertisement";
import Features from "../../../components/client/Features/Features";
import FeaturedCategories from "../../../components/client/featuredCategories/FeaturedCategories";
import BestSellingProducts from "../../../components/client/BestSellingProducts/BestSellingProducts";
import LatestProducts from "../../../components/client/LatestProducts/LatestProducts";
import Footer from "../../../components/common/Footer/Footer";
import SearchBar from "../../../components/common/SearchBar/SearchBar";
import UserSettings from "../../../components/common/UserSettings/UserSettings";

// Import Image
import img1 from "../../../assets/images/istockphoto-1170073824-612x612.jpg";

function ClientHomePage() {
  const data = {
    arbAdTitle: "كل ما تريده ي مكان واحد",
    adTitle: "Everything You Want On One Place",
    arbAdSubTitle: "اكتشف منتجات مذهلة، وأسعار رائعة،",
    adSubTitle: "Discover Amazing Products, great price and",
    arbCom: "وتجربة تسوق سلسة.",
    com: "a seamless shopping experience",
    imgSrc: img1,
  };

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
