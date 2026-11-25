// Import Style
import {
  ProductStaticsContainerClass,
  avrageRateClass,
  starsContainerClass,
  rateBarsContainerClass,
  rateBarContainerClass,
  barContainerClass,
  goldenStarClass,
  notGoldenStarClass,
} from "./ProductStatics.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";
import { averageRating } from "../../../hooks/useProductReviews";
import { calcCountForEachRate } from "../../../utils/helpers";

function ProductStatics({ reviews }) {
  const { lang } = useContext(AuthContext);
  const average = averageRating(reviews);
  const counts = calcCountForEachRate(reviews);

  const rates = [5, 4, 3, 2, 1];
  return (
    <>
      <div className={ProductStaticsContainerClass}>
        <div className={avrageRateClass}>
          <h1>{average.toFixed(1)}</h1>
          <div className={starsContainerClass}>
            {Array.from({ length: 5 }, (_, i) => (
              <i
                key={i}
                className={`fa-solid fa-star ${
                  i < Math.round(average)
                    ? goldenStarClass
                    : notGoldenStarClass
                }`}
              ></i>
            ))}
          </div>
          <div>
            {lang === "arb"
              ? `بناء على استطلاع ${reviews.length} آراء`
              : `Based On ${reviews.length} Reviews`}
          </div>
        </div>
        <div className={rateBarsContainerClass}>
          {rates.map((rate, index) => (
            <div className={rateBarContainerClass} key={index}>
              <div>{rate}</div>
              <div>
                <i className="fa-solid fa-star"></i>
              </div>
              <div className={barContainerClass}>
                <div
                  style={{
                    width: `${reviews.length ? (counts[rate - 1] / reviews.length) * 100 : 0}%`,
                  }}
                ></div>
              </div>
              <div>{counts[rate - 1]}</div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

export default ProductStatics;
