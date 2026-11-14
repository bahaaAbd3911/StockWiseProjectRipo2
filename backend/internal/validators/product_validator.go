package validators

import (
	"errors"
	"strings"

	"ecommerce-backend/internal/models"
)

func ValidateProduct(product models.CreateProductRequest) error {
	if strings.TrimSpace(product.Name) == "" {
		return errors.New("product name is required")
	}

	if strings.TrimSpace(product.NameAr) == "" {
		return errors.New("Arabic product name is required")
	}

	if product.Price < 0 {
		return errors.New("price cannot be negative")
	}

	if product.DiscountValue < 0 || product.DiscountValue > 100 {
		return errors.New("discount must be between 0 and 100")
	}

	if product.Stock < 0 {
		return errors.New("stock cannot be negative")
	}

	return nil
}
func ValidateUpdateProduct(product models.UpdateProductRequest) error {
	if strings.TrimSpace(product.Name) == "" {
		return errors.New("product name is required")
	}

	if strings.TrimSpace(product.NameAr) == "" {
		return errors.New("Arabic product name is required")
	}

	if product.Price < 0 {
		return errors.New("price cannot be negative")
	}

	if product.DiscountValue < 0 || product.DiscountValue > 100 {
		return errors.New("discount must be between 0 and 100")
	}

	if product.Stock < 0 {
		return errors.New("stock cannot be negative")
	}

	return nil
}
