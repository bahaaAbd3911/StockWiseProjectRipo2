package handlers

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
)

// ProductHandler serves products, categories, images and colors.
type ProductHandler struct {
	products *services.ProductService
}

func NewProductHandler(products *services.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

// parseProductFilter reads ?category= &search= &sort= &in_stock=true &limit= &offset=
func parseProductFilter(r *http.Request) (models.ProductFilter, error) {
	q := r.URL.Query()
	limit, offset, err := httpx.Pagination(r)
	if err != nil {
		return models.ProductFilter{}, err
	}
	f := models.ProductFilter{
		Category:    strings.TrimSpace(q.Get("category")),
		Search:      strings.TrimSpace(q.Get("search")),
		Sort:        strings.TrimSpace(q.Get("sort")),
		InStockOnly: q.Get("in_stock") == "true",
		Limit:       limit,
		Offset:      offset,
	}
	if utf8.RuneCountInString(f.Search) > 100 {
		return f, apperr.Invalid("search text is too long (maximum 100 characters)")
	}
	return f, nil
}

// List: GET /api/products (active products only)
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) error {
	f, err := parseProductFilter(r)
	if err != nil {
		return err
	}
	page, err := h.products.List(r.Context(), f)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// AdminList: GET /api/admin/products (includes inactive products)
func (h *ProductHandler) AdminList(w http.ResponseWriter, r *http.Request) error {
	f, err := parseProductFilter(r)
	if err != nil {
		return err
	}
	f.IncludeInactive = true
	page, err := h.products.List(r.Context(), f)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, page)
}

// Get: GET /api/products/{id} -> product with images and colors
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) error {
	return h.get(w, r, false)
}

// AdminGet: GET /api/admin/products/{id} (also inactive products)
func (h *ProductHandler) AdminGet(w http.ResponseWriter, r *http.Request) error {
	return h.get(w, r, true)
}

func (h *ProductHandler) get(w http.ResponseWriter, r *http.Request, includeInactive bool) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	p, err := h.products.Get(r.Context(), id, includeInactive)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, p)
}

// Categories: GET /api/categories
func (h *ProductHandler) Categories(w http.ResponseWriter, r *http.Request) error {
	cats, err := h.products.Categories(r.Context())
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, cats)
}

// Create: POST /api/products (admin)
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var in models.ProductInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	p, err := h.products.Create(r.Context(), in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, p)
}

// Update: PUT /api/products/{id} (admin)
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.ProductInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	p, err := h.products.Update(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, p)
}

// Delete: DELETE /api/products/{id} (admin, soft delete)
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.products.Deactivate(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "product deactivated")
}

// ---------- images ----------

// Images: GET /api/products/{id}/images
func (h *ProductHandler) Images(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	imgs, err := h.products.Images(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, imgs)
}

// AddImage: POST /api/products/{id}/images (admin)
func (h *ProductHandler) AddImage(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.ProductImageInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	img, err := h.products.AddImage(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, img)
}

// SetPrimaryImage: PUT /api/product-images/{id}/primary (admin)
func (h *ProductHandler) SetPrimaryImage(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.products.SetPrimaryImage(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "primary image updated")
}

// DeleteImage: DELETE /api/product-images/{id} (admin)
func (h *ProductHandler) DeleteImage(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.products.DeleteImage(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "image deleted")
}

// ---------- colors ----------

// Colors: GET /api/products/{id}/colors
func (h *ProductHandler) Colors(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	colors, err := h.products.Colors(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, colors)
}

// AddColor: POST /api/products/{id}/colors (admin)
func (h *ProductHandler) AddColor(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	var in models.ProductColorInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		return err
	}
	c, err := h.products.AddColor(r.Context(), id, in)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, c)
}

// DeleteColor: DELETE /api/product-colors/{id} (admin)
func (h *ProductHandler) DeleteColor(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	if err := h.products.DeleteColor(r.Context(), id); err != nil {
		return err
	}
	return httpx.Message(w, http.StatusOK, "color deleted")
}
