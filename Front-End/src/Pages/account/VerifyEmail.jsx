import { useContext, useEffect, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router";

import { verifyEmail } from "../../api/auth";
import { AuthContext } from "../../context/AuthContext";
import { useLang } from "../../hooks/useLang";
import AccountCard from "./AccountCard";
import styles from "./AccountPages.module.css";

// Opened from the sign-up email: /verify-email?token=...
// Confirms the address once; checkout needs a confirmed email.
export default function VerifyEmail() {
  const [params] = useSearchParams();
  const token = params.get("token") ?? "";
  const { tr } = useLang();
  const { isSign, refreshProfile } = useContext(AuthContext);
  const [state, setState] = useState({ status: token ? "working" : "error", message: "" });
  const sent = useRef(false); // React may run effects twice in development; a link works only once

  useEffect(() => {
    if (!token || sent.current) return;
    sent.current = true;
    verifyEmail(token)
      .then(() => {
        setState({ status: "done", message: "" });
        if (isSign) refreshProfile?.();
      })
      .catch((err) => setState({ status: "error", message: err.message }));
  }, [token, isSign, refreshProfile]);

  return (
    <AccountCard icon="fa-solid fa-envelope-circle-check" title={tr("Confirm your email", "تأكيد البريد الإلكتروني")}>
      {state.status === "working" && <p className={styles.text}>{tr("Confirming…", "جارٍ التأكيد…")}</p>}

      {state.status === "done" && (
        <>
          <p className={styles.success}>
            {tr("Thank you, your email address is confirmed.", "شكراً لك، تم تأكيد بريدك الإلكتروني.")}
          </p>
          <p className={styles.text}>{tr("You can now place orders.", "يمكنك الآن تقديم الطلبات.")}</p>
          {!isSign && (
            <Link to="/signIn" className={styles.back}>
              {tr("Sign in", "تسجيل الدخول")}
            </Link>
          )}
        </>
      )}

      {state.status === "error" && (
        <>
          <p className={styles.error}>
            {state.message ||
              tr("This confirmation link is missing its code.", "رابط التأكيد غير مكتمل.")}
          </p>
          <p className={styles.text}>
            {tr(
              "You can ask for a new link from your profile page.",
              "يمكنك طلب رابط جديد من صفحة ملفك الشخصي.",
            )}
          </p>
        </>
      )}
    </AccountCard>
  );
}
