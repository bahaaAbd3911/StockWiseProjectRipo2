import {
  UserProfileMainInfoContainerClass,
  userDataOrdersPageContainerClass,
  darkUserProfileMainInfoContainerClass,
} from "../../../components/client/UserProfileMainInfo/UserProfileMainInfo.module.css";

// Import Images

// Import Motion
import { motion } from "motion/react";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";
function UserOrdersMainInfo() {
  const { lang, dark } = useContext(AuthContext);
  return (
    <>
      <motion.div
        className={`${UserProfileMainInfoContainerClass} ${dark && darkUserProfileMainInfoContainerClass}`}
        initial={{ opacity: 0, y: -170 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.4 }}
      >
        <div>
          <div className={userDataOrdersPageContainerClass}>
            <div>
              <div>{lang === "arb" ? "طلباتي" : "My Orders"}</div>
            </div>
            <div>
              {lang === "arb"
                ? "تابع طلباتك بسهولة"
                : "Track Your Orders Easily"}
            </div>
          </div>
        </div>
      </motion.div>
    </>
  );
}

export default UserOrdersMainInfo;
