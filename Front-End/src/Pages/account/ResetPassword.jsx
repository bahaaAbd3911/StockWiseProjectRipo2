import { useState } from "react";
import { Link, useSearchParams } from "react-router";

import { forgotPassword, resetPassword } from "../../api/auth";
import { Button, FormField } from "../../components/admin/ui";
import { useLang } from "../../hooks/useLang";
import AccountCard from "./AccountCard";
import styles from "./AccountPages.module.css";

const MIN_PASSWORD = 8; // same rule as the backend

// Two steps on one page:
//   /reset-password            -> "Forgot password": enter the email, a link is sent
//   /reset-password?token=...  -> opened from that email: choose a new password
export default function ResetPassword() {
  const [params] = useSearchParams();
  const token = params.get("token");
  return token ? <NewPassword token={token} /> : <RequestLink />;
}

function RequestLink() {
  const { tr } = useLang();
  const [email, setEmail] = useState("");
  const [state, setState] = useState({ status: "idle", message: "" });

  const submit = async (e) => {
    e.preventDefault();
    if (!email.trim()) return setState({ status: "error", message: tr("Enter your email", "أدخل بريدك الإلكتروني") });
    setState({ status: "working", message: "" });
    try {
      await forgotPassword(email.trim());
      setState({ status: "done", message: "" });
    } catch (err) {
      setState({ status: "error", message: err.message });
    }
  };

  return (
    <AccountCard icon="fa-solid fa-key" title={tr("Forgot your password?", "نسيت كلمة المرور؟")}>
      {state.status === "done" ? (
        <p className={styles.success}>
          {tr(
            "If an account exists for this email, we sent a link to reset the password. Check your inbox (and spam).",
            "إذا كان هناك حساب بهذا البريد، فقد أرسلنا رابطاً لإعادة تعيين كلمة المرور. تحقق من بريدك (والرسائل غير المرغوبة).",
          )}
        </p>
      ) : (
        <form className={styles.form} onSubmit={submit} noValidate>
          <p className={styles.text}>
            {tr("Enter your email and we will send you a link.", "أدخل بريدك الإلكتروني وسنرسل لك رابطاً.")}
          </p>
          <FormField
            label={tr("Email", "البريد الإلكتروني")}
            name="email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={state.status === "error" ? state.message : ""}
          />
          <Button type="submit" disabled={state.status === "working"}>
            {tr("Send the link", "إرسال الرابط")}
          </Button>
        </form>
      )}
    </AccountCard>
  );
}

function NewPassword({ token }) {
  const { tr } = useLang();
  const [form, setForm] = useState({ password: "", confirm: "" });
  const [state, setState] = useState({ status: "idle", message: "" });

  const submit = async (e) => {
    e.preventDefault();
    if (form.password.length < MIN_PASSWORD)
      return setState({ status: "error", message: tr("At least 8 characters", "8 أحرف على الأقل") });
    if (form.password !== form.confirm)
      return setState({ status: "error", message: tr("The passwords do not match", "كلمتا المرور غير متطابقتين") });
    setState({ status: "working", message: "" });
    try {
      await resetPassword(token, form.password);
      setState({ status: "done", message: "" });
    } catch (err) {
      setState({ status: "error", message: err.message });
    }
  };

  if (state.status === "done") {
    return (
      <AccountCard icon="fa-solid fa-circle-check" title={tr("Password changed", "تم تغيير كلمة المرور")}>
        <p className={styles.text}>
          {tr(
            "For your safety you were signed out on every device. Sign in with your new password.",
            "لحمايتك تم تسجيل خروجك من جميع الأجهزة. سجّل الدخول بكلمة المرور الجديدة.",
          )}
        </p>
        <Link to="/signIn" className={styles.back}>
          {tr("Sign in", "تسجيل الدخول")}
        </Link>
      </AccountCard>
    );
  }

  return (
    <AccountCard icon="fa-solid fa-lock" title={tr("Choose a new password", "اختر كلمة مرور جديدة")}>
      <form className={styles.form} onSubmit={submit} noValidate>
        <FormField
          label={tr("New password", "كلمة المرور الجديدة")}
          name="password"
          type="password"
          autoComplete="new-password"
          value={form.password}
          onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
        />
        <FormField
          label={tr("Repeat the new password", "أعد كتابة كلمة المرور")}
          name="confirm"
          type="password"
          autoComplete="new-password"
          value={form.confirm}
          onChange={(e) => setForm((f) => ({ ...f, confirm: e.target.value }))}
        />
        {state.status === "error" && <p className={styles.error}>{state.message}</p>}
        <Button type="submit" disabled={state.status === "working"}>
          {tr("Save the new password", "حفظ كلمة المرور")}
        </Button>
        {state.status === "error" && (
          <Link to="/reset-password" className={styles.back}>
            {tr("Ask for a new link", "اطلب رابطاً جديداً")}
          </Link>
        )}
      </form>
    </AccountCard>
  );
}
