import { api, dataUrlToFile, isDataUrl, tokenStore, upload } from "./client";

export const login = (email, password) =>
  api("/api/login", { method: "POST", body: { email, password } });

// Public sign-up: the backend always creates a customer and emails a link to
// confirm the address. Registration does not log in, so we log in right after.
// A profile photo picked in the form (a data URL) is uploaded once logged in.
export const register = async ({ profile_image_url: photo, ...data }) => {
  await api("/api/register", { method: "POST", body: data });
  const session = await login(data.email, data.password);
  if (isDataUrl(photo)) {
    tokenStore.set(session.token); // the upload needs the new login
    try {
      session.user = await uploadProfileImage(photo);
    } catch {
      // The account exists; the photo can be added later from the profile.
    }
  }
  return session;
};

export const getProfile = () => api("/api/profile");

// profile_image_url may be: an existing URL (kept), "" (removed) or a newly
// picked photo as a data URL (uploaded through the image endpoint, which
// checks the file and strips location data).
export const updateProfile = async ({ profile_image_url: photo, ...data }) => {
  let url = photo || null;
  if (isDataUrl(photo)) {
    // Upload first: it replaces (and deletes) the previous uploaded photo.
    url = (await uploadProfileImage(photo)).profile_image_url;
  } else if (!photo) {
    await api("/api/profile/image", { method: "DELETE" }); // also deletes an uploaded file
  }
  return api("/api/profile", { method: "PUT", body: { ...data, profile_image_url: url } });
};

async function uploadProfileImage(dataUrl) {
  return upload("/api/profile/image", await dataUrlToFile(dataUrl, "profile"));
}

export const changePassword = (currentPassword, newPassword) =>
  api("/api/profile/password", {
    method: "PUT",
    body: { current_password: currentPassword, new_password: newPassword },
  });

// Email confirmation and password reset (pages linked from the emails).
export const verifyEmail = (token) => api("/api/email/verify", { method: "POST", body: { token } });
export const resendVerification = () => api("/api/email/verify/resend", { method: "POST" });
export const forgotPassword = (email) => api("/api/password/forgot", { method: "POST", body: { email } });
export const resetPassword = (token, newPassword) =>
  api("/api/password/reset", { method: "POST", body: { token, new_password: newPassword } });
