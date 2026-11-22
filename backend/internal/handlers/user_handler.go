package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type UserHandler struct {
	users *services.UserService
}

func NewUserHandler(users *services.UserService) *UserHandler { return &UserHandler{users: users} }

// Register: POST /api/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req models.RegisterRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	user, err := h.users.Register(r.Context(), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, user)
}

// Login: POST /api/login -> {user, token, expires_at}
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req models.LoginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	res, err := h.users.Login(r.Context(), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, res)
}

// Profile: GET /api/profile
func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) error {
	user, err := h.users.Profile(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, user)
}

// UpdateProfile: PUT /api/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) error {
	var req models.UpdateProfileRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	user, err := h.users.UpdateProfile(r.Context(), auth.UserID(r.Context()), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, user)
}

// ChangePassword: PUT /api/profile/password -> {message, token, expires_at}
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) error {
	var req models.ChangePasswordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	res, err := h.users.ChangePassword(r.Context(), auth.UserID(r.Context()), req)
	if err != nil {
		return err
	}
	// Other devices are now logged out; this response carries the new token
	// for the current device.
	return httpx.JSON(w, http.StatusOK, res)
}

// LogoutEverywhere: POST /api/logout-all - revokes every token of the user,
// including the one used for this request.
//
// For a normal "log out" on one device the frontend simply deletes its token.
func (h *UserHandler) LogoutEverywhere(w http.ResponseWriter, r *http.Request) error {
	if err := h.users.LogoutEverywhere(r.Context(), auth.UserID(r.Context())); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "logged out on all devices")
}

// ForgotPassword: POST /api/password/forgot {email}
// Always answers the same way, whether or not the email has an account.
func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) error {
	var req models.ForgotPasswordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	if err := h.users.RequestPasswordReset(r.Context(), req, httpx.ClientIP(r.Context())); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, services.ForgotPasswordMessage)
}

// ResetPassword: POST /api/password/reset {token, new_password}
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	var req models.ResetPasswordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	if err := h.users.ResetPassword(r.Context(), req); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "password has been reset; please log in with your new password")
}

// Deactivate: DELETE /api/profile
func (h *UserHandler) Deactivate(w http.ResponseWriter, r *http.Request) error {
	if err := h.users.Deactivate(r.Context(), auth.UserID(r.Context())); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "account deactivated")
}
