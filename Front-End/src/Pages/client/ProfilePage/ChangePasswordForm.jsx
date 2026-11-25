import { useState } from "react";

import { changePassword } from "../../../api/auth";
import { Button, FormField } from "../../../components/admin/ui";
import { useApiAction } from "../../../hooks/useApiAction";
import { isStrongPassword } from "../../../utils/validators";

export default function ChangePasswordForm({ tr, onDone }) {
  const { run, busy } = useApiAction();
  const [form, setForm] = useState({ current: "", next: "", confirm: "" });
  const [errors, setErrors] = useState({});
  const set = (key) => (e) => setForm((prev) => ({ ...prev, [key]: e.target.value }));

  const submit = async (e) => {
    e.preventDefault();
    const found = {};
    if (!form.current) found.current = tr("Required", "مطلوب");
    if (!isStrongPassword(form.next))
      found.next = tr("8+ characters with upper case, lower case and a number", "8 أحرف على الأقل مع حرف كبير وصغير ورقم");
    if (form.confirm !== form.next) found.confirm = tr("Passwords do not match", "كلمتا المرور غير متطابقتين");
    setErrors(found);
    if (Object.keys(found).length) return;

    const ok = await run(() => changePassword(form.current, form.next), ["Password changed", "تم تغيير كلمة المرور"]);
    if (ok) onDone();
  };

  const field = (key, label, autoComplete) => (
    <FormField
      label={label}
      name={key}
      type="password"
      autoComplete={autoComplete}
      value={form[key]}
      error={errors[key]}
      onChange={set(key)}
    />
  );

  return (
    <form onSubmit={submit} noValidate>
      {field("current", tr("Current password", "كلمة المرور الحالية"), "current-password")}
      {field("next", tr("New password", "كلمة المرور الجديدة"), "new-password")}
      {field("confirm", tr("Confirm new password", "تأكيد كلمة المرور الجديدة"), "new-password")}
      <Button type="submit" icon="fa-solid fa-lock" disabled={busy}>
        {tr("Change Password", "تغيير كلمة المرور")}
      </Button>
    </form>
  );
}
