package handlers

import (
	"net/http"

	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/services"
)

type DashboardHandler struct {
	dashboard *services.DashboardService
}

func NewDashboardHandler(d *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: d}
}

// Get: GET /api/admin/dashboard?days=30 (1-366)
func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) error {
	days, err := httpx.QueryInt(r, "days", services.DefaultDashboardDays, 1, services.MaxDashboardDays)
	if err != nil {
		return err
	}
	d, err := h.dashboard.Get(r.Context(), days)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, d)
}
