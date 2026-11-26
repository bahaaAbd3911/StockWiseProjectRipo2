// Import Style
import {
  SearchBarContainerClass,
  SearchBarLangIsArbContainerClass,
  inputContainerClass,
  cartAndFavoriteContainerClass,
  logoContainerClass,
  searchAndIconsContainerClass,
  logoDarkClass,
  cartAndFavoriteContainerDarkClass,
  profileBtnContainerClass,
  heartLoveClass,
  shoppingCartClass,
} from "./SearchBar.module.css";

// Import Hooks
import { useContext } from "react";
import { useShop } from "../../../hooks/useShop";
import { AuthContext } from "../../../context/AuthContext";
import { useLocation, useNavigate } from "react-router";
import { useBrand } from "../../../hooks/useBrand";
import BrandMark from "../BrandLogo/BrandMark";

// Import Motion
import { motion } from "motion/react";

// Import Images

function SearchBar() {
  const { dark, lang, setSearchFilters, searchFilters, openProfileTab, isSign, user } =
    useContext(AuthContext);
  const { cartCount, favorites, requireSignIn } = useShop();
  // Cart / favorites / profile all live in the profile page tabs.
  const openTab = (index) => {
    if (index !== 0 && !requireSignIn()) return;
    openProfileTab(index);
    navigate("/client/profile");
  };
  const brand = useBrand();
  let navigate = useNavigate();
  const { pathname } = useLocation();
  return (
    <>
      <motion.div
        className={`${SearchBarContainerClass} ${lang === "arb" ? SearchBarLangIsArbContainerClass : ""}`}
        initial={{ opacity: 0, x: -100 }}
        whileInView={{ opacity: 1, x: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.5 }}
      >
        <div
          className={logoContainerClass}
          onClick={() => {
            navigate("/client/home");
          }}
        >
          <div>
            <BrandMark />
          </div>
          <div className={dark ? logoDarkClass : ""}>
            {brand.first}
            <span>{brand.rest}</span>
          </div>
        </div>
        <div className={searchAndIconsContainerClass}>
          <div className={inputContainerClass}>
            <input
              type="search"
              aria-label={lang === "arb" ? "بحث" : "Search"}
              value={searchFilters.productName}
              placeholder={
                lang === "arb"
                  ? "أبحث عن المنتج الذي تريده"
                  : "Search for your product"
              }
              onChange={(e) => {
                setSearchFilters((prev) => ({
                  ...prev,
                  productName: e.target.value,
                }));
                // Results are shown on the products page.
                if (pathname !== "/client/products") navigate("/client/products");
              }}
            />
          </div>
          <div
            className={`${cartAndFavoriteContainerClass}  ${dark ? cartAndFavoriteContainerDarkClass : ""}`}
          >
            <div
              className={heartLoveClass}
              role="button"
              title={lang === "arb" ? "العروض" : "Offers"}
              aria-label={lang === "arb" ? "العروض" : "Offers"}
              onClick={() => navigate("/client/offers")}
            >
              <i className="fa-solid fa-tags"></i>
            </div>
            <div
              className={heartLoveClass}
              role="button"
              title={lang === "arb" ? "الباقات" : "Bundles"}
              aria-label={lang === "arb" ? "الباقات" : "Bundles"}
              onClick={() => navigate("/client/bundles")}
            >
              <i className="fa-solid fa-boxes-stacked"></i>
            </div>
            <div
              className={heartLoveClass}
              role="button"
              aria-label={lang === "arb" ? "المفضلة" : "Favorites"}
              onClick={() => openTab(3)}
            >
              <i className="fa-solid fa-heart"></i>
              {favorites.length > 0 && <span>{favorites.length}</span>}
            </div>
            <div
              className={shoppingCartClass}
              role="button"
              aria-label={lang === "arb" ? "سلة التسوق" : "Cart"}
              onClick={() => openTab(2)}
            >
              <i className="fa-solid fa-cart-shopping"></i>
              {cartCount > 0 && <span>{cartCount}</span>}
            </div>
            <div
              className={profileBtnContainerClass}
              role="button"
              aria-label={lang === "arb" ? "الصفحة الشخصية" : "Profile"}
              onClick={() => (isSign ? openTab(0) : navigate("/signIn"))}
            >
              {user?.image ? (
                <img src={user.image} alt="" />
              ) : (
                <i className={`fa-solid ${isSign ? "fa-user" : "fa-right-to-bracket"}`}></i>
              )}
            </div>
          </div>
        </div>
      </motion.div>
    </>
  );
}

export default SearchBar;
