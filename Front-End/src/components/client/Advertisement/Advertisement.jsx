// Import Style
import {
  AdvertisementContainerClass,
  infoContainerClass,
  imageClass,
  arbicAdvertisementContainerClass,
  darkAdvertisementContainerClass,
} from "./Advertisement.module.css";

// Import Image

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";
import { useNavigate } from "react-router";

// Import Motion
import { motion } from "motion/react";

function Advertisement({ data, value }) {
  const { dark, lang } = useContext(AuthContext);
  let navigate = useNavigate();

  return (
    <>
      <motion.div
        className={`${AdvertisementContainerClass} ${lang === "arb" ? arbicAdvertisementContainerClass : ""} ${dark && darkAdvertisementContainerClass}`}
        initial={{ opacity: 0, y: 70 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.4 }}
      >
        <div className={infoContainerClass}>
          <h1>{lang === "arb" ? data.arbAdTitle : data.adTitle}</h1>
          <div>
            {lang === "arb" ? data.arbAdSubTitle : data.adSubTitle}
            <span>
              <br /> {lang === "arb" ? data.arbCom : data.com}
            </span>
          </div>
          <button
            onClick={() => {
              if (!value) {
                navigate("/client/products");
              }
            }}
          >
            <span>
              {lang === "arb" ? data.arbButtonText || "تسوق الآن" : data.buttonText || "Shop Now"}
            </span>
            <span>
              {lang === "arb" ? (
                <i className="fa-solid fa-arrow-left"></i>
              ) : (
                <i className="fa-solid fa-arrow-right-long"></i>
              )}
            </span>
          </button>
        </div>
        <div className={imageClass}>
          <img src={data.imgSrc} />
        </div>
      </motion.div>
    </>
  );
}

export default Advertisement;
