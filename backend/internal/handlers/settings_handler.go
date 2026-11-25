package handlers

import (
	"net/http"

	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type SettingsHandler struct {
	settings *services.SettingsService
}

func NewSettingsHandler(s *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{settings: s}
}

// Get: GET /api/settings (public: every visitor needs the name, logo and colors)
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) error {
	s, err := h.settings.Get(r.Context())
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, s)
}

// Save: PUT /api/admin/settings {storeName, logo, colors}
func (h *SettingsHandler) Save(w http.ResponseWriter, r *http.Request) error {
	var in models.StoreSettings
	if err := httpx.DecodeLargeJSON(w, r, &in, services.MaxSettingsBody); err != nil {
		return err
	}
	s, err := h.settings.Save(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, s)
}
