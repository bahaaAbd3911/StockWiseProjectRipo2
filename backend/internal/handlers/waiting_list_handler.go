package handlers

import (
	"net/http"
	"strconv"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type WaitingListHandler struct {
	items *services.WaitingListService
}

func NewWaitingListHandler(items *services.WaitingListService) *WaitingListHandler {
	return &WaitingListHandler{items: items}
}

// Join: POST /api/waiting-list {product_id, quantity, selected_color}
func (h *WaitingListHandler) Join(w http.ResponseWriter, r *http.Request) error {
	var req models.CreateWaitingListRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	item, err := h.items.Join(r.Context(), auth.UserID(r.Context()), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, item)
}

// ListMine: GET /api/waiting-list
func (h *WaitingListHandler) ListMine(w http.ResponseWriter, r *http.Request) error {
	list, err := h.items.ListMine(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, list)
}

// CancelMine: DELETE /api/waiting-list/{id}
func (h *WaitingListHandler) CancelMine(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.items.CancelMine(r.Context(), auth.UserID(r.Context()), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "waiting-list request cancelled")
}

// AdminList: GET /api/admin/waiting-list?status=&product_id=&limit=&offset=
func (h *WaitingListHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	f := models.WaitingListFilter{Status: r.URL.Query().Get("status"), Limit: limit, Offset: offset}
	if raw := r.URL.Query().Get("product_id"); raw != "" {
		f.ProductID, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || f.ProductID <= 0 {
			return apperr.Invalid("query parameter \"product_id\" must be a positive number")
		}
	}
	page, err := h.items.List(r.Context(), f)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// AdminUpdateStatus: PUT /api/admin/waiting-list/{id}/status {status}
func (h *WaitingListHandler) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.UpdateWaitingListStatusRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	item, err := h.items.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, item)
}
