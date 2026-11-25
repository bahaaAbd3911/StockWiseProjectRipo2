import { useState } from "react";
import { motion } from "motion/react";

import { useLang } from "../../../hooks/useLang";
import {
  ProductDetailsContainerClass,
  arbProductDetailsContainerClass,
  productDataContainerClass,
  productDataContainerClassX,
} from "./ProductDetails.module.css";

import ImagesContainer from "./ImagesContainer";
import StockCase from "./StockCase";
import ProductHeading from "./ProductHeading";
import ProductDiscount from "./ProductDiscount";
import ProductDescription from "./ProductDescription";
import OrderedProduct from "./OrderedProduct";
import ChooseProductColor from "./ChooseProductColor";
import ProductButtons from "./ProductButtons";

// The parent remounts this per product (key), so the state below starts fresh.
function ProductDetails({ productData }) {
  const { isAr } = useLang();
  const [activeImage, setActiveImage] = useState(productData.images[0] ?? productData.image);
  const [quantity, setQuantity] = useState(1);
  const [color, setColor] = useState(productData.availableColors?.[0]?.name ?? null);

  return (
    <motion.div
      className={`${ProductDetailsContainerClass} ${isAr ? arbProductDetailsContainerClass : ""}`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
    >
      <ImagesContainer productData={productData} setActicveImage={setActiveImage} activeImage={activeImage} />
      <div className={productDataContainerClass}>
        <StockCase productData={productData} />
        <div className={productDataContainerClassX}>
          <ProductHeading productData={productData} />
          <ProductDiscount productData={productData} />
          <ProductDescription productData={productData} />
          {!productData.outOfStock && (
            <OrderedProduct productData={productData} quantity={quantity} setQuantity={setQuantity} />
          )}
          <ChooseProductColor productData={productData} color={color} setColor={setColor} />
          <ProductButtons productData={productData} quantity={quantity} color={color} />
        </div>
      </div>
    </motion.div>
  );
}

export default ProductDetails;
