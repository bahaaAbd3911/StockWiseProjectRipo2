import { useNavigate } from "react-router";

import { useLang } from "../../../hooks/useLang";
import styles from "./SignInRequired.module.css";

// Inline message shown in place of pages that need an account (cart, favorites, profile…).
export default function SignInRequired({ text }) {
  const navigate = useNavigate();
  const { tr, dir } = useLang();

  return (
    <div className={styles.box} dir={dir} role="alert">
      <i className="fa-solid fa-lock"></i>
      <h2>{tr("You must sign in first", "يجب عليك تسجيل الدخول أولاً")}</h2>
      <p>{text ?? tr("Sign in to use your cart, favorites and orders.", "سجّل الدخول لاستخدام السلة والمفضلة والطلبات.")}</p>
      <div className={styles.actions}>
        <button className={styles.primary} onClick={() => navigate("/signIn")}>
          {tr("Sign In", "تسجيل الدخول")}
        </button>
        <button className={styles.secondary} onClick={() => navigate("/signUp")}>
          {tr("Create Account", "إنشاء حساب")}
        </button>
      </div>
    </div>
  );
}
