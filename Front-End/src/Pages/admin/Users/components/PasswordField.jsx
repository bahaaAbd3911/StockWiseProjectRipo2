import { useState } from "react";
import { FormField, IconButton, inputClass } from "../../../../components/admin/ui";

export default function PasswordField({ value, error, hint, onChange, tr }) {
  const [visible, setVisible] = useState(false);

  return (
    <FormField label={tr("Password", "كلمة المرور")} error={error} hint={hint}>
      <div style={{ display: "flex", gap: 8 }}>
        <input
          id="userPassword"
          className={inputClass}
          type={visible ? "text" : "password"}
          autoComplete="new-password"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        <IconButton
          icon={`fa-solid ${visible ? "fa-eye-slash" : "fa-eye"}`}
          label={visible ? tr("Hide password", "إخفاء") : tr("Show password", "إظهار")}
          onClick={() => setVisible((v) => !v)}
        />
      </div>
    </FormField>
  );
}
