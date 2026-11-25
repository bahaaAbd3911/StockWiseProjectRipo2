import { useState } from "react";

import { Button, FormField, Modal, SelectInput, formGridClass } from "../../../../components/admin/ui";
import { USER_ROLES, USER_STATUSES } from "../../../../constants/userRoles";
import { pickLabel } from "../../../../utils/format";
import { initialUserForm, toUser, validateUser } from "../userForm";
import PasswordField from "./PasswordField";

const toOptions = (map, isAr) => Object.entries(map).map(([value, label]) => ({ value, label: pickLabel(label, isAr) }));

export default function UserFormModal({ user, users, isAr, tr, onSave, onClose }) {
  const [form, setForm] = useState(() => initialUserForm(user));
  const [errors, setErrors] = useState({});
  const update = (key, value) => setForm((prev) => ({ ...prev, [key]: value }));
  const field = (key) => ({ name: key, value: form[key], error: errors[key], onChange: (e) => update(key, e.target.value) });

  const submit = (e) => {
    e.preventDefault();
    const found = validateUser(form, users, user?.userID, tr);
    setErrors(found);
    if (!Object.keys(found).length) onSave(toUser(form));
  };

  return (
    <Modal
      title={user ? tr("Edit User", "تعديل المستخدم") : tr("Add User", "إضافة مستخدم")}
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>{tr("Cancel", "إلغاء")}</Button>
          <Button type="submit" form="userForm" icon="fa-solid fa-check">{tr("Save User", "حفظ المستخدم")}</Button>
        </>
      }
    >
      <form id="userForm" onSubmit={submit} noValidate>
        <div className={formGridClass}>
          {user && <FormField label={tr("User ID", "معرف المستخدم")} name="userID" value={form.userID} readOnly disabled />}
          <FormField label={tr("Full name", "الاسم الكامل")} {...field("userFullName")} />
          <FormField label={tr("Email", "البريد الإلكتروني")} type="email" {...field("userEmail")} />
          <FormField label={tr("Phone", "الهاتف")} placeholder="07XXXXXXXX" {...field("phone")} />
        </div>

        <PasswordField
          value={form.userPassword}
          error={errors.userPassword}
          hint={user ? tr("Leave blank to keep the current password", "اتركه فارغًا للإبقاء على كلمة المرور الحالية") : null}
          tr={tr}
          onChange={(v) => update("userPassword", v)}
        />

        <div className={formGridClass}>
          <FormField label={tr("Role", "الدور")}>
            <SelectInput value={form.role} onChange={(v) => update("role", v)} options={toOptions(USER_ROLES, isAr)} />
          </FormField>
          <FormField label={tr("Status", "الحالة")}>
            <SelectInput value={form.status} onChange={(v) => update("status", v)} options={toOptions(USER_STATUSES, isAr)} />
          </FormField>
        </div>
      </form>
    </Modal>
  );
}
