import { useContext, useState } from "react";

import { AuthContext } from "../../../context/AuthContext";
import { Button, FormField, formGridClass } from "../../../components/admin/ui";
import { readImage } from "../../../data/authForms";
import { useApiAction } from "../../../hooks/useApiAction";

const MAX_IMAGE_BYTES = 500 * 1024;
const PHONE_RE = /^(?:\+962|00962|0)?7[789]\d{7}$/;

// Saves name, phone and photo through PUT /api/profile. (Email can't be changed here.)
export default function ProfileDetailsForm({ tr, onDone }) {
  const { user, updateProfile } = useContext(AuthContext);
  const { run, busy } = useApiAction();
  const [form, setForm] = useState({
    firstName: user.firstName,
    lastName: user.lastName,
    phone: user.phone,
    image: user.image,
  });
  const [errors, setErrors] = useState({});
  const set = (key) => (e) => setForm((prev) => ({ ...prev, [key]: e.target.value }));

  const pickImage = async (file) => {
    if (!file) return;
    if (file.size > MAX_IMAGE_BYTES) return setErrors({ image: tr("Image must be under 500 KB", "يجب أن تكون الصورة أقل من 500 كيلوبايت") });
    setErrors({});
    setForm((prev) => ({ ...prev, image: "" }));
    const image = await readImage(file);
    setForm((prev) => ({ ...prev, image }));
  };

  const submit = async (e) => {
    e.preventDefault();
    const found = {};
    if (!form.firstName.trim()) found.firstName = tr("Required", "مطلوب");
    if (!form.lastName.trim()) found.lastName = tr("Required", "مطلوب");
    if (!PHONE_RE.test(form.phone.replace(/[\s-]/g, ""))) found.phone = tr("Enter a valid Jordanian phone number", "أدخل رقم هاتف أردني صحيح");
    setErrors(found);
    if (Object.keys(found).length) return;

    const ok = await run(
      () =>
        updateProfile({
          first_name: form.firstName.trim(),
          last_name: form.lastName.trim(),
          phone: form.phone.trim(),
          profile_image_url: form.image,
        }),
      ["Profile updated", "تم تحديث الملف الشخصي"],
    );
    if (ok) onDone();
  };

  return (
    <form onSubmit={submit} noValidate>
      <div className={formGridClass}>
        <FormField label={tr("First name", "الاسم الأول")} name="firstName" value={form.firstName} error={errors.firstName} onChange={set("firstName")} />
        <FormField label={tr("Last name", "اسم العائلة")} name="lastName" value={form.lastName} error={errors.lastName} onChange={set("lastName")} />
      </div>
      <FormField label={tr("Phone", "الهاتف")} name="phone" type="tel" value={form.phone} error={errors.phone} onChange={set("phone")} />
      <FormField label={tr("Email (cannot be changed)", "البريد الإلكتروني (لا يمكن تغييره)")} name="email" value={user.userEmail} disabled readOnly />
      <FormField label={tr("Profile photo", "الصورة الشخصية")} error={errors.image}>
        <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
          {form.image && <img src={form.image} alt="" style={{ width: 56, height: 56, borderRadius: "50%", objectFit: "cover" }} />}
          <input type="file" accept="image/*" onChange={(e) => pickImage(e.target.files[0])} />
          {form.image && (
            <Button variant="ghost" onClick={() => setForm((prev) => ({ ...prev, image: "" }))}>
              {tr("Remove", "إزالة")}
            </Button>
          )}
        </div>
      </FormField>
      <Button type="submit" icon="fa-solid fa-check" disabled={busy}>
        {tr("Save Changes", "حفظ التغييرات")}
      </Button>
    </form>
  );
}
