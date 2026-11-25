import { useContext } from "react";

import { AuthContext } from "../../context/AuthContext";
import { useLang } from "../../hooks/useLang";
import AccessNotice from "../../components/common/AccessNotice/AccessNotice";

// Only signed-in admins get past this; everyone else sees an error message.
// (The backend enforces the same rule on every admin API call.)
export default function AdminGuard({ children }) {
  const { ready, isSign, isAdmin } = useContext(AuthContext);
  const { tr } = useLang();

  if (!ready) {
    return <AccessNotice loading title={tr("Checking your account…", "جارٍ التحقق من حسابك…")} />;
  }

  if (!isSign) {
    return (
      <AccessNotice
        icon="fa-solid fa-lock"
        title={tr("Sign in required", "يجب تسجيل الدخول")}
        text={tr(
          "You must sign in with an administrator account to open the admin panel.",
          "يجب عليك تسجيل الدخول بحساب مدير لفتح لوحة التحكم.",
        )}
        actionLabel={tr("Sign In", "تسجيل الدخول")}
        actionTo="/signIn"
      />
    );
  }

  if (!isAdmin) {
    return (
      <AccessNotice
        icon="fa-solid fa-ban"
        title={tr("Access denied", "تم رفض الوصول")}
        text={tr(
          "Your account does not have permission to open the admin panel.",
          "حسابك لا يملك صلاحية فتح لوحة التحكم.",
        )}
      />
    );
  }

  return children;
}
