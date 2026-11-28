package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type OrderHandler struct {
	orders *services.OrderService
}

func NewOrderHandler(orders *services.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

// Checkout: POST /api/orders {address_id?, payment_method?}
// address_id omitted = default address; payment_method omitted = cash_on_delivery
func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) error {
	var req models.CreateOrderRequest
	if err := httpx.DecodeOptionalJSON(w, r, &req); err != nil { // an empty body is allowed
		return err
	}
	order, err := h.orders.Checkout(r.Context(), auth.UserID(r.Context()), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, order)
}

// ListMine: GET /api/orders
func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) error {
	orders, err := h.orders.ListMine(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, orders)
}

// GetMine: GET /api/orders/{id}
func (h *OrderHandler) GetMine(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	order, err := h.orders.GetMine(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, order)
}

// CancelMine: POST /api/orders/{id}/cancel
func (h *OrderHandler) CancelMine(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	order, err := h.orders.CancelMine(r.Context(), auth.UserID(r.Context()), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, order)
}

// AdminList: GET /api/admin/orders?status=&payment_method=&limit=&offset=
func (h *OrderHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	page, err := h.orders.List(r.Context(), models.OrderFilter{
		Status: r.URL.Query().Get("status"), PaymentMethod: r.URL.Query().Get("payment_method"),
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// AdminGet: GET /api/admin/orders/{id}
func (h *OrderHandler) AdminGet(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	order, err := h.orders.Get(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, order)
}

// AdminSetNote: PUT /api/admin/orders/{id}/note {note}
func (h *OrderHandler) AdminSetNote(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.UpdateOrderNoteRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	order, err := h.orders.SetNote(r.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, order)
}

// AdminUpdateStatus: PUT /api/admin/orders/{id}/status {status, payment_collected?}
func (h *OrderHandler) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.UpdateOrderStatusRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	order, err := h.orders.UpdateStatus(r.Context(), auth.UserID(r.Context()), id, req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, order)
}
