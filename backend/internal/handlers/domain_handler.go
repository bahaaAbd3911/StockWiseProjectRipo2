package handlers

import (
	"net/http"

	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
	"ecommerce-backend/internal/tenant"
)

// DomainHandler: a store admin's custom domains, and the HTTPS check Caddy
// calls before issuing a certificate.
type DomainHandler struct {
	domains      *services.DomainService
	resolver     *tenant.Resolver
	platformHost string
}

func NewDomainHandler(d *services.DomainService, resolver *tenant.Resolver, platformHost string) *DomainHandler {
	return &DomainHandler{domains: d, resolver: resolver, platformHost: platformHost}
}

// List: GET /api/admin/domains
func (h *DomainHandler) List(w http.ResponseWriter, r *http.Request) error {
	list, err := h.domains.List(r.Context())
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, list)
}

// Add: POST /api/admin/domains {domain}
func (h *DomainHandler) Add(w http.ResponseWriter, r *http.Request) error {
	var req models.AddDomainRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	d, err := h.domains.Add(r.Context(), req)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, d)
}

func (h *DomainHandler) withID(fn func(*http.Request, int64) (any, error)) httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := httpx.PathID(r, "id")
		if err != nil {
			return err
		}
		out, err := fn(r, id)
		if err != nil {
			return err
		}
		return httpx.JSON(w, http.StatusOK, out)
	}
}

// Verify: POST /api/admin/domains/{id}/verify
func (h *DomainHandler) Verify(w http.ResponseWriter, r *http.Request) error {
	return h.withID(func(r *http.Request, id int64) (any, error) { return h.domains.Verify(r.Context(), id) })(w, r)
}

// SetPrimary: POST /api/admin/domains/{id}/primary
func (h *DomainHandler) SetPrimary(w http.ResponseWriter, r *http.Request) error {
	return h.withID(func(r *http.Request, id int64) (any, error) { return h.domains.SetPrimary(r.Context(), id) })(w, r)
}

// Delete: DELETE /api/admin/domains/{id}
func (h *DomainHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.domains.Delete(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "domain removed")
}

// TLSCheck: GET /internal/tls-check?domain=... (Caddy's on-demand TLS "ask").
// 200 = issue a certificate for this host, 404 = refuse. Only hosts the
// platform really serves pass, so nobody can make the server request
// certificates for random names.
func (h *DomainHandler) TLSCheck(w http.ResponseWriter, r *http.Request) error {
	ok, err := h.resolver.IsServed(r.Context(), r.URL.Query().Get("domain"), h.platformHost)
	if err != nil {
		return err
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	return nil
}
