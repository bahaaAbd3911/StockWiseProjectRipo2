package handlers

import (
	"net/http"
	"strconv"

	"ecommerce-backend/internal/services"
)

// ExportHandler serves the admin's CSV downloads.
type ExportHandler struct {
	exports *services.ExportService
}

func NewExportHandler(exports *services.ExportService) *ExportHandler {
	return &ExportHandler{exports: exports}
}

// Orders: GET /api/admin/exports/orders.csv?from=2026-11-01&to=2026-11-30&status=delivered&payment_method=pay_later
func (h *ExportHandler) Orders(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	file, err := h.exports.Orders(r.Context(), q.Get("from"), q.Get("to"), q.Get("status"), q.Get("payment_method"))
	if err != nil {
		return err
	}
	return writeCSVFile(w, file)
}

// Customers: GET /api/admin/exports/customers.csv
func (h *ExportHandler) Customers(w http.ResponseWriter, r *http.Request) error {
	file, err := h.exports.Customers(r.Context())
	if err != nil {
		return err
	}
	return writeCSVFile(w, file)
}

func writeCSVFile(w http.ResponseWriter, file services.ExportFile) error {
	h := w.Header()
	h.Set("Content-Type", "text/csv; charset=utf-8")
	h.Set("Content-Disposition", `attachment; filename="`+file.Name+`"`)
	h.Set("Content-Length", strconv.Itoa(len(file.Data)))
	h.Set("Cache-Control", "no-store") // customer data: never kept in a cache
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(file.Data)
	return err
}
