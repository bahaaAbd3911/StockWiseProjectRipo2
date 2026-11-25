// Field definitions for the sign-in, sign-up and edit-profile forms.
const field = (index, placeHolder, arbPlaceHolder, iconClass, type, extra = {}) => ({
  placeHolder,
  arbPlaceHolder,
  data: "",
  flag: false,
  errorMsg: "",
  arbErrorMsg: "",
  index,
  iconClass,
  type,
  ...extra,
});

export const signInFields = () => [
  field(0, "User Email", "عنوان البريد الإلكتروني", "fa-solid fa-envelope", "email", { inputType: "email" }),
  field(1, "Password", "كلمة المرور", "fa-solid fa-lock", "password", { inputType: "password" }),
];

// Same order is used by SignUpValidation / EditFormValidation (index 0..5).
export const signUpFields = (initial = {}) => [
  field(0, "Enter Full Name", "أدخل الاسم الكامل", "fa-solid fa-user", "text", { data: initial.name ?? "" }),
  field(1, "User Email", "عنوان البريد الإلكتروني", "fa-solid fa-envelope", "text", { data: initial.email ?? "" }),
  field(2, "Password", "كلمة المرور", "fa-solid fa-lock", "password"),
  field(3, "Confirm Password", "تأكيد كلمة المرور", "fa-solid fa-lock", "password"),
  field(4, "Enter Your Phone Number", "أدخل رقم هاتفك", "fa-solid fa-phone", "tel", { data: initial.phone ?? "" }),
  field(5, "Upload Your Personal Image (optional)", "أضف صورتك الشخصية (اختياري)", "fa-solid fa-camera", "file"),
];

// Marks one field with a server-side error message.
export const withFieldError = (fields, index, errorMsg, arbErrorMsg = errorMsg) =>
  fields.map((f) => (f.index === index ? { ...f, flag: true, errorMsg, arbErrorMsg } : f));

// Reads an image file as a data URL (stored as the profile image).
export const readImage = (file) =>
  new Promise((resolve, reject) => {
    if (!file) return resolve("");
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
