package handlers

import (
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

type DebtHandler struct {
	debts *services.DebtService
}

func NewDebtHandler(debts *services.DebtService) *DebtHandler { return &DebtHandler{debts: debts} }

func (h *DebtHandler) statement(w http.ResponseWriter, r *http.Request, userID int64) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	st, err := h.debts.Statement(r.Context(), userID, limit, offset)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, st)
}

// Mine: GET /api/debt - the logged-in customer's balance, credit and history
func (h *DebtHandler) Mine(w http.ResponseWriter, r *http.Request) error {
	return h.statement(w, r, auth.UserID(r.Context()))
}

// AdminList: GET /api/admin/debts?outstanding=true
func (h *DebtHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return err
	}
	page, err := h.debts.ListAccounts(r.Context(), models.DebtAccountFilter{
		OutstandingOnly: r.URL.Query().Get("outstanding") == "true", Limit: limit, Offset: offset})
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// AdminStatement: GET /api/admin/customers/{id}/debt
func (h *DebtHandler) AdminStatement(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	return h.statement(w, r, id)
}

// SetCreditLimit: PUT /api/admin/customers/{id}/credit-limit {credit_limit}
func (h *DebtHandler) SetCreditLimit(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.SetCreditLimitRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	acc, err := h.debts.SetCreditLimit(r.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, acc)
}

// RecordPayment: POST /api/admin/customers/{id}/debt/payments {amount, note?, order_id?}
func (h *DebtHandler) RecordPayment(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.RecordPaymentRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	st, err := h.debts.RecordPayment(r.Context(), auth.UserID(r.Context()), id, req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, st)
}

// RecordAdjustment: POST /api/admin/customers/{id}/debt/adjustments {amount, note, order_id?}
func (h *DebtHandler) RecordAdjustment(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var req models.RecordAdjustmentRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	st, err := h.debts.RecordAdjustment(r.Context(), auth.UserID(r.Context()), id, req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, st)
}
