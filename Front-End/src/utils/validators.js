export const isValidEmail = (value) =>
  /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(String(value).trim());

// At least 8 characters with upper case, lower case and a number.
export const isStrongPassword = (value) =>
  /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/.test(String(value));

export const isHexColor = (value) =>
  /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(String(value).trim());
