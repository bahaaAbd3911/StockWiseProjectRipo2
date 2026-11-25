// Import Style
import {
  LoadingPageContainerClass,
  shapeClass,
  headingContainerClass,
  mainTitleContainerClass,
  subTitleClass,
  topLeftWaveClass,
  topRightShopping,
  bottomLeftShoppingBag,
  bottomRightWave,
  mainContentContainerClass,
  loadingBarContainerClass,
  loadingBarClass,
  loadClass,
  featuresContainerClass,
  featureClass,
  shoppingCartImgContainer,
} from "./LoadingPage.module.css";

// Import Images
import bottom_left_shopping_bag from "../../assets/images/Loading/bottom_left_shopping_bag_watermark.png";
import bottom_right_wave from "../../assets/images/Loading/bottom_right_wave_background.png";
import shopping_cart from "../../assets/images/Loading/shopping_cart_illustration.png";
import top_left_wave from "../../assets/images/Loading/top_left_wave_background.png";
import top_right_shopping from "../../assets/images/Loading/top_right_shopping_bag_watermark.png";

// Import Hooks
import { useContext, useEffect } from "react";
import { useNavigate } from "react-router";

// Import Theme
import { StoreContext } from "../../context/StoreContext";
import { COLOR_TOKENS } from "../../data/storeSettings";

function LoadingPage() {
  let navigate = useNavigate();
  const { storeSettings } = useContext(StoreContext);

  // The loading page always uses the light (morning) colors, whatever the selected mode.
  const lightTheme = Object.fromEntries(
    COLOR_TOKENS.map(({ key, css }) => [css, storeSettings.colors.light[key]]),
  );

  useEffect(() => {
    setTimeout(() => {
      navigate("/client/home");
    }, 4000);
  });
  return (
    <>
      <div className={LoadingPageContainerClass} style={lightTheme}>
        <div className={`${shapeClass} ${topLeftWaveClass}`}>
          <img src={top_left_wave} />
        </div>
        <div className={`${shapeClass}  ${topRightShopping}`}>
          <img src={top_right_shopping} />
        </div>
        <div className={`${shapeClass}  ${bottomLeftShoppingBag}`}>
          <img src={bottom_left_shopping_bag} />
        </div>
        <div className={`${shapeClass}  ${bottomRightWave}`}>
          <img src={bottom_right_wave} />
        </div>
        <div className={mainContentContainerClass}>
          <div className={headingContainerClass}>
            <div className={mainTitleContainerClass}>
              <h1>
                <div>
                  <i className="fa-solid fa-basket-shopping"></i>
                </div>
                <div>
                  <span>Stock</span>
                  <span> Wise</span>
                </div>
              </h1>
            </div>
            <div className={subTitleClass}>
              Everything You Want In One Place
            </div>
          </div>
          <div className={shoppingCartImgContainer}>
            <img src={shopping_cart} />
          </div>
          <div className={loadingBarContainerClass}>
            <div>Loading Your Shopping Experience</div>
            <div className={loadingBarClass}>
              <div className={loadClass}></div>
            </div>
          </div>
          <div className={featuresContainerClass}>
            <div className={featureClass}></div>
          </div>
        </div>
      </div>
    </>
  );
}

export default LoadingPage;
