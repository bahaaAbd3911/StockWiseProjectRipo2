import { Link, useParams } from "react-router";

import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import UserSettings from "../../common/UserSettings/UserSettings";
import SearchBar from "../../common/SearchBar/SearchBar";
import Footer from "../../common/Footer/Footer";
import CatalogStatus from "../CatalogStatus/CatalogStatus";
import ProductDetails from "../ProductDetails/ProductDetails";
import SimilarProducts from "../SimilarProducts/SimilarProducts";
import CommentOnProducts from "../CommentOnProducts/CommentOnProducts";
import { notFoundClass } from "../ProductDetails/ProductDetails.module.css";

function ProductBody({ productID }) {
  const { products } = useCatalog();
  const { tr } = useLang();
  const product = products.find((p) => String(p.id) === String(productID));

  if (!product) {
    return (
      <div className={notFoundClass}>
        <h2>{tr("Product not found", "المنتج غير موجود")}</h2>
        <Link to="/client/products">{tr("Browse all products", "تصفح كل المنتجات")}</Link>
      </div>
    );
  }

  return (
    <>
      <ProductDetails key={product.id} productData={product} />
      <CommentOnProducts key={`reviews-${product.id}`} productID={product.id} />
      <SimilarProducts product={product} />
    </>
  );
}

function ProductDetailsPage() {
  const { productID } = useParams();

  return (
    <div>
      <UserSettings />
      <SearchBar value={1} />
      <CatalogStatus>
        <ProductBody productID={productID} />
      </CatalogStatus>
      <Footer />
    </div>
  );
}

export default ProductDetailsPage;
