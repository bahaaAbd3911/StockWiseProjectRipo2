package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type CouponHandler struct {
	coupons *services.CouponService
}

func NewCouponHandler(coupons *services.CouponService) *CouponHandler {
	return &CouponHandler{coupons: coupons}
}

// Apply: POST /api/cart/coupon {code} -> the updated cart
func (h *CouponHandler) Apply(w http.ResponseWriter, r *http.Request) error {
	var req models.ApplyCouponRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	cart, err := h.coupons.ApplyToCart(r.Context(), auth.UserID(r.Context()), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// Remove: DELETE /api/cart/coupon -> the updated cart
func (h *CouponHandler) Remove(w http.ResponseWriter, r *http.Request) error {
	cart, err := h.coupons.RemoveFromCart(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cart)
}

// AdminList: GET /api/admin/coupons?status=&search=&limit=&offset=
func (h *CouponHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	page, err := h.coupons.List(r.Context(), models.CouponFilter{
		Status: r.URL.Query().Get("status"), Search: r.URL.Query().Get("search"), Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// AdminGet: GET /api/admin/coupons/{id}
func (h *CouponHandler) AdminGet(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	c, err := h.coupons.Get(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, c)
}

// Create: POST /api/admin/coupons
func (h *CouponHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var in models.CouponInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	c, err := h.coupons.Create(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, c)
}

// Update: PUT /api/admin/coupons/{id}
func (h *CouponHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.CouponInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	c, err := h.coupons.Update(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, c)
}

// Delete: DELETE /api/admin/coupons/{id}
func (h *CouponHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.coupons.Delete(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "coupon deleted")
}
