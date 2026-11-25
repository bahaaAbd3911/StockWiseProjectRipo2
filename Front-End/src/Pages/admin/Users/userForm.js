import { isStrongPassword, isValidEmail } from "../../../utils/validators";

// userID is assigned by the database, so it is shown but not editable.
export const initialUserForm = (user) => ({
  userID: user?.userID ?? "",
  userFullName: user?.userFullName ?? "",
  userEmail: user?.userEmail ?? "",
  userPassword: "",
  phone: user?.phone ?? "",
  role: user?.role ?? "customer",
  status: user?.status ?? "active",
});

export function validateUser(form, users, originalId, tr) {
  const errors = {};
  const others = users.filter((u) => u.userID !== originalId);
  const isNew = originalId === undefined;

  if (form.userFullName.trim().split(/\s+/).length < 2)
    errors.userFullName = tr("Enter first and last name", "أدخل الاسم الأول والأخير");

  if (!isValidEmail(form.userEmail)) errors.userEmail = tr("Enter a valid email", "أدخل بريدًا صحيحًا");
  else if (others.some((u) => u.userEmail.toLowerCase() === form.userEmail.trim().toLowerCase()))
    errors.userEmail = tr("This email is already used", "هذا البريد مستخدم");

  // Password is required for new users; when editing, blank keeps the current one.
  if ((isNew || form.userPassword) && !isStrongPassword(form.userPassword))
    errors.userPassword = tr(
      "8+ characters with upper case, lower case and a number",
      "8 أحرف على الأقل مع حرف كبير وصغير ورقم",
    );

  if (form.phone && !/^07\d{8}$/.test(form.phone))
    errors.phone = tr("Phone must look like 07XXXXXXXX", "رقم الهاتف بصيغة 07XXXXXXXX");

  return errors;
}

export const toUser = (form) => ({
  ...form,
  userFullName: form.userFullName.trim(),
  userEmail: form.userEmail.trim(),
});
