import { useState } from "react";

import { Modal } from "../../../components/admin/ui";
import { useLang } from "../../../hooks/useLang";
import ProfileDetailsForm from "./ProfileDetailsForm";
import ChangePasswordForm from "./ChangePasswordForm";

const tabStyle = (active) => ({
  flex: 1,
  padding: "10px",
  border: "none",
  borderRadius: 10,
  font: "inherit",
  fontWeight: 600,
  cursor: "pointer",
  background: active ? "var(--color-primary)" : "transparent",
  color: active ? "var(--color-on-primary)" : "var(--color-text)",
});

export default function ProfileEditModal({ onClose }) {
  const { tr } = useLang();
  const [tab, setTab] = useState("details");

  return (
    <Modal title={tr("Edit Profile", "تعديل الملف الشخصي")} onClose={onClose}>
      <div
        role="tablist"
        style={{ display: "flex", gap: 6, padding: 4, marginBottom: 18, borderRadius: 12, background: "var(--color-primary-soft)" }}
      >
        <button role="tab" aria-selected={tab === "details"} style={tabStyle(tab === "details")} onClick={() => setTab("details")}>
          {tr("Personal Info", "المعلومات الشخصية")}
        </button>
        <button role="tab" aria-selected={tab === "password"} style={tabStyle(tab === "password")} onClick={() => setTab("password")}>
          {tr("Password", "كلمة المرور")}
        </button>
      </div>

      {tab === "details" ? <ProfileDetailsForm tr={tr} onDone={onClose} /> : <ChangePasswordForm tr={tr} onDone={onClose} />}
    </Modal>
  );
}
