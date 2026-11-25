// Import Style
import {
  featuresContainerClass,
  arbFeaturesContainerClass,
} from "./features.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";
import { useStoreContent } from "../../../hooks/useStoreContent";

// Import Motion
import { motion } from "motion/react";

// Import Component
import Feature from "./Feature";

function Features() {
  const { lang } = useContext(AuthContext);

  // The store's own tiles (Store Settings), or the defaults.
  const features = useStoreContent().features.map((f) => ({
    icon: f.icon,
    name: f.name || f.nameAr,
    arbName: f.nameAr || f.name,
    comment: f.comment || f.commentAr,
    arbComment: f.commentAr || f.comment,
  }));
  return (
    <>
      <motion.div
        className={`${featuresContainerClass}   ${lang === "arb" ? arbFeaturesContainerClass : ""}`}
        initial={{ opacity: 0, y: 70 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.4 }}
      >
        {features.map((feture, index) => (
          <Feature feture={feture} index={index} key={index} />
        ))}
      </motion.div>
    </>
  );
}
export default Features;
