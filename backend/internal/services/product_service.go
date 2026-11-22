package services

import (
	"context"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// ProductService manages products and their images/colors.
type ProductService struct {
	products *repository.ProductRepository
	catalog  *repository.CatalogRepository
}

func NewProductService(products *repository.ProductRepository, catalog *repository.CatalogRepository) *ProductService {
	return &ProductService{products: products, catalog: catalog}
}

func (s *ProductService) List(ctx context.Context, f models.ProductFilter) (models.Page[models.Product], error) {
	if f.Sort == "" {
		f.Sort = "newest"
	}
	if !repository.IsValidSort(f.Sort) {
		return models.Page[models.Product]{}, apperr.Validation(map[string]string{
			"sort": "must be one of: newest, price_asc, price_desc, name"})
	}
	return s.products.List(ctx, f)
}

// Get returns a product with its images and colors. Customers cannot see
// inactive products; admins (includeInactive) can.
func (s *ProductService) Get(ctx context.Context, id int64, includeInactive bool) (models.ProductDetail, error) {
	p, err := s.products.Get(ctx, id, includeInactive)
	if err != nil {
		return models.ProductDetail{}, err
	}
	images, err := s.catalog.ListImages(ctx, id)
	if err != nil {
		return models.ProductDetail{}, err
	}
	colors, err := s.catalog.ListColors(ctx, id)
	if err != nil {
		return models.ProductDetail{}, err
	}
	return models.ProductDetail{Product: p, Images: images, Colors: colors}, nil
}

func (s *ProductService) Create(ctx context.Context, in models.ProductInput) (models.ProductDetail, error) {
	if err := validators.ValidateProduct(&in); err != nil {
		return models.ProductDetail{}, err
	}
	id, err := s.products.Create(ctx, in)
	if err != nil {
		return models.ProductDetail{}, err
	}
	return s.Get(ctx, id, true)
}

func (s *ProductService) Update(ctx context.Context, id int64, in models.ProductInput) (models.ProductDetail, error) {
	if err := validators.ValidateProduct(&in); err != nil {
		return models.ProductDetail{}, err
	}
	if err := s.products.Update(ctx, id, in); err != nil {
		return models.ProductDetail{}, err
	}
	return s.Get(ctx, id, true)
}

func (s *ProductService) Deactivate(ctx context.Context, id int64) error {
	return s.products.Deactivate(ctx, id)
}

func (s *ProductService) Categories(ctx context.Context) ([]models.Category, error) {
	return s.products.Categories(ctx)
}

// ---------- images & colors ----------

// requireProduct returns 404 unless the product exists (and, for public
// requests, is active).
func (s *ProductService) requireProduct(ctx context.Context, id int64, activeOnly bool) error {
	ok, err := s.catalog.ProductExists(ctx, id, activeOnly)
	if err != nil {
		return err
	}
	if !ok {
		return apperr.NotFound("product not found")
	}
	return nil
}

func (s *ProductService) Images(ctx context.Context, productID int64) ([]models.ProductImage, error) {
	if err := s.requireProduct(ctx, productID, true); err != nil {
		return nil, err
	}
	return s.catalog.ListImages(ctx, productID)
}

func (s *ProductService) Colors(ctx context.Context, productID int64) ([]models.ProductColor, error) {
	if err := s.requireProduct(ctx, productID, true); err != nil {
		return nil, err
	}
	return s.catalog.ListColors(ctx, productID)
}

func (s *ProductService) AddImage(ctx context.Context, productID int64, in models.ProductImageInput) (models.ProductImage, error) {
	if err := validators.ValidateProductImage(&in); err != nil {
		return models.ProductImage{}, err
	}
	return s.catalog.AddImage(ctx, productID, in)
}

func (s *ProductService) SetPrimaryImage(ctx context.Context, imageID int64) error {
	return s.catalog.SetPrimaryImage(ctx, imageID)
}

func (s *ProductService) DeleteImage(ctx context.Context, imageID int64) error {
	return s.catalog.DeleteImage(ctx, imageID)
}

func (s *ProductService) AddColor(ctx context.Context, productID int64, in models.ProductColorInput) (models.ProductColor, error) {
	if err := validators.ValidateProductColor(&in); err != nil {
		return models.ProductColor{}, err
	}
	if err := s.requireProduct(ctx, productID, false); err != nil {
		return models.ProductColor{}, err
	}
	return s.catalog.AddColor(ctx, productID, in)
}

func (s *ProductService) DeleteColor(ctx context.Context, colorID int64) error {
	return s.catalog.DeleteColor(ctx, colorID)
}
