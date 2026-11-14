package handlers

import (
	"database/sql"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/services"
	"ecommerce-backend/internal/validators"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type ProductHandler struct {
	Service *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{
		Service: service,
	}
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.Service.GetAllProducts()
	if err != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(products)
}
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	// Get the product ID from the URL
	idString := r.PathValue("id")

	// Convert the ID from string to int64
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Ask the service to retrieve the product
	product, err := h.Service.GetProductByID(id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the product as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(product)
}
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var request models.CreateProductRequest

	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate product data
	err = validators.ValidateProduct(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send product data to service
	product, err := h.Service.CreateProduct(request)
	if err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	// Return created product
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(product)

}
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Get product ID from URL
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Decode JSON request
	var product models.UpdateProductRequest

	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate product
	err = validators.ValidateUpdateProduct(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update product
	updatedProduct, err := h.Service.UpdateProduct(id, product)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	// Return updated product
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProduct)
}
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Get product ID from URL
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Delete product
	err = h.Service.DeleteProduct(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"message": "Product deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}
