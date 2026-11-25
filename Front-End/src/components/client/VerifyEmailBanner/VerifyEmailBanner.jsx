import { useContext, useState } from "react";

import { resendVerification } from "../../../api/auth";
import { AuthContext } from "../../../context/AuthContext";
import { useLang } from "../../../hooks/useLang";
import { useToast } from "../../../hooks/useToast";
import styles from "./VerifyEmailBanner.module.css";

// Shown while the signed-in customer has not confirmed their email. Ordering
// needs a confirmed email, so this explains why and offers a new link.
export default function VerifyEmailBanner() {
  const { user, isAdmin, refreshProfile } = useContext(AuthContext);
  const { tr, dir } = useLang();
  const toast = useToast();
  const [busy, setBusy] = useState(false);

  if (!user || user.emailVerified || isAdmin) return null;

  const resend = async () => {
    setBusy(true);
    try {
      const res = await resendVerification();
      toast.success(res.message, tr(res.message, "أرسلنا رابط تأكيد جديد إلى بريدك الإلكتروني"));
      refreshProfile?.();
    } catch (err) {
      toast.error(err.message, err.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className={styles.banner} dir={dir} role="status">
      <i className="fa-solid fa-envelope" />
      <span>
        {tr(
          `Please confirm your email address (${user.userEmail}) to place orders. Check your inbox for our link.`,
          `يرجى تأكيد بريدك الإلكتروني (${user.userEmail}) لتتمكن من الطلب. تحقق من بريدك الوارد.`,
        )}
      </span>
      <button type="button" onClick={resend} disabled={busy}>
        {tr("Send the link again", "إعادة إرسال الرابط")}
      </button>
    </div>
  );
}
