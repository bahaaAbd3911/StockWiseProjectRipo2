package handlers

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/services"
	"ecommerce-backend/internal/validators"
)

// UploadHandler receives image files as multipart/form-data (what an HTML
// <input type="file"> or JavaScript FormData sends).
type UploadHandler struct {
	products *services.ProductService
	users    *services.UserService
	maxBytes int64
}

func NewUploadHandler(products *services.ProductService, users *services.UserService, maxBytes int64) *UploadHandler {
	return &UploadHandler{products: products, users: users, maxBytes: maxBytes}
}

// form is the parsed upload: the file in field "image" plus small text fields.
type form struct {
	image  []byte
	fields map[string]string
}

// readForm streams the request. The file is read into memory with a hard
// limit, so an oversized upload is refused without touching the disk.
func (h *UploadHandler) readForm(w http.ResponseWriter, r *http.Request) (form, error) {
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "multipart/form-data" {
		return form{}, validators.FieldError("image", "send the file as multipart/form-data in a field called image")
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+64<<10) // + room for the form's own headers
	mr, err := r.MultipartReader()
	if err != nil {
		return form{}, validators.FieldError("image", "invalid multipart form")
	}
	f := form{fields: map[string]string{}}
	tooBig := validators.FieldError("image", fmt.Sprintf("must be at most %d MB", h.maxBytes>>20))
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return form{}, tooBig
		}
		if err != nil {
			return form{}, validators.FieldError("image", "invalid multipart form")
		}
		name := part.FormName()
		switch {
		case name == "image" && part.FileName() != "":
			data, err := io.ReadAll(io.LimitReader(part, h.maxBytes+1))
			if errors.As(err, &maxErr) || int64(len(data)) > h.maxBytes {
				return form{}, tooBig
			}
			if err != nil {
				return form{}, validators.FieldError("image", "upload interrupted")
			}
			f.image = data
		case name != "":
			v, _ := io.ReadAll(io.LimitReader(part, 1024))
			f.fields[name] = strings.TrimSpace(string(v))
		}
		part.Close()
	}
	if len(f.image) == 0 {
		return form{}, validators.FieldError("image", "is required (a JPEG, PNG or WebP file)")
	}
	return f, nil
}

// ProductImage: POST /api/admin/products/{id}/images/upload (admin)
// Form fields: image (file), is_primary (true/false, optional), display_order (optional).
func (h *UploadHandler) ProductImage(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		return err
	}
	f, err := h.readForm(w, r)
	if err != nil {
		return err
	}
	isPrimary := false
	if v := f.fields["is_primary"]; v != "" {
		if isPrimary, err = strconv.ParseBool(v); err != nil {
			return validators.FieldError("is_primary", "must be true or false")
		}
	}
	order := 0
	if v := f.fields["display_order"]; v != "" {
		if order, err = strconv.Atoi(v); err != nil {
			return validators.FieldError("display_order", "must be a whole number")
		}
	}
	img, err := h.products.UploadImage(r.Context(), id, f.image, isPrimary, order)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, img)
}

// ProfileImage: POST /api/profile/image (logged in). Form field: image (file).
func (h *UploadHandler) ProfileImage(w http.ResponseWriter, r *http.Request) error {
	f, err := h.readForm(w, r)
	if err != nil {
		return err
	}
	u, err := h.users.UploadProfileImage(r.Context(), auth.UserID(r.Context()), f.image)
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, u)
}

// DeleteProfileImage: DELETE /api/profile/image (logged in)
func (h *UploadHandler) DeleteProfileImage(w http.ResponseWriter, r *http.Request) error {
	u, err := h.users.DeleteProfileImage(r.Context(), auth.UserID(r.Context()))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusOK, u)
}
