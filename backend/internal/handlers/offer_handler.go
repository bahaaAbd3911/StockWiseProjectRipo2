package handlers

import (
	"net/http"

	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type OfferHandler struct {
	offers *services.OfferService
}

func NewOfferHandler(offers *services.OfferService) *OfferHandler {
	return &OfferHandler{offers: offers}
}

// ListLive: GET /api/offers - offers running right now
func (h *OfferHandler) ListLive(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	page, err := h.offers.ListLive(r.Context(), limit, offset)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// GetLive: GET /api/offers/{id} -> {offer, products: page}
func (h *OfferHandler) GetLive(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	offer, products, err := h.offers.GetLive(r.Context(), id, limit, offset)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, map[string]any{"offer": offer, "products": products})
}

// AdminList: GET /api/admin/offers?status=live|scheduled|expired|disabled
func (h *OfferHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	page, err := h.offers.List(r.Context(), models.OfferFilter{
		Status: r.URL.Query().Get("status"), Limit: limit, Offset: offset})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// AdminGet: GET /api/admin/offers/{id}
func (h *OfferHandler) AdminGet(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	o, err := h.offers.Get(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, o)
}

// Create: POST /api/admin/offers
func (h *OfferHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var in models.OfferInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	o, err := h.offers.Create(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, o)
}

// Update: PUT /api/admin/offers/{id}
func (h *OfferHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.OfferInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	o, err := h.offers.Update(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, o)
}

// Delete: DELETE /api/admin/offers/{id}
func (h *OfferHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.offers.Delete(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "offer deleted")
}
