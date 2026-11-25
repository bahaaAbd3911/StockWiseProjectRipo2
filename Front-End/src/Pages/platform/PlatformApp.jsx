import { useContext, useEffect, useState } from "react";

import { AuthContext } from "../../context/AuthContext";
import { platformLogin, platformMe, platformToken } from "../../api/platform";
import { Button, Card, FormField } from "../../components/admin/ui";
import Message from "../../components/client/Message/Message";
import UserSettings from "../../components/common/UserSettings/UserSettings";
import { useLang } from "../../hooks/useLang";
import PlatformStores from "./PlatformStores";
import "../../layouts/AdminLayout/adminTheme.css";
import styles from "./Platform.module.css";

// What the platform admin host (admin.<platform domain>) shows instead of a
// shop: the super admin's login and the list of stores. Uses the same admin
// UI kit and theme as a store's admin panel.
export default function PlatformApp() {
  const { dark, dir, tr } = useLang();
  const { message, setMessage } = useContext(AuthContext);
  const [admin, setAdmin] = useState(null);
  const [ready, setReady] = useState(() => !platformToken.get());

  useEffect(() => {
    if (!platformToken.get()) return;
    platformMe()
      .then(setAdmin)
      .catch(() => platformToken.clear())
      .finally(() => setReady(true));
  }, []);

  useEffect(() => {
    if (!message.isShow) return;
    const timer = setTimeout(() => setMessage({ isShow: false, msgTxt: "", arbMsgTxt: "", msgBGColor: "" }), message.time);
    return () => clearTimeout(timer);
  }, [message, setMessage]);

  const signOut = () => {
    platformToken.clear();
    setAdmin(null);
  };

  return (
    <div className={`adminTheme ${dark ? "adminThemeDark" : ""} ${styles.shell}`} dir={dir}>
      {message.isShow && (
        <Message mesgTxt={message.msgTxt} arbMsgTxt={message.arbMsgTxt} msgBGColor={message.msgBGColor} reminder={message.reminder} />
      )}
      <UserSettings />
      <main className={styles.content}>
        {!ready ? null : admin ? (
          <PlatformStores admin={admin} onSignOut={signOut} onUnauthorized={signOut} />
        ) : (
          <Login onDone={setAdmin} tr={tr} />
        )}
      </main>
    </div>
  );
}

function Login({ onDone, tr }) {
  const [form, setForm] = useState({ email: "", password: "" });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async (e) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      onDone(await platformLogin(form.email.trim(), form.password));
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card className={styles.login}>
      <h1>
        <i className="fa-solid fa-shield-halved" /> {tr("Platform admin", "مدير المنصة")}
      </h1>
      <p className={styles.muted}>{tr("Sign in to manage the stores.", "سجّل الدخول لإدارة المتاجر.")}</p>
      <form onSubmit={submit} noValidate className={styles.form}>
        <FormField label={tr("Email", "البريد الإلكتروني")} name="email" type="email" autoComplete="username"
          value={form.email} onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))} />
        <FormField label={tr("Password", "كلمة المرور")} name="password" type="password" autoComplete="current-password"
          value={form.password} onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))} error={error} />
        <Button type="submit" icon="fa-solid fa-right-to-bracket" disabled={busy}>
          {tr("Sign in", "تسجيل الدخول")}
        </Button>
      </form>
    </Card>
  );
}
