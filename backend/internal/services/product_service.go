package services

import (
	"context"
	"log/slog"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/images"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/storage"
	"ecommerce-backend/internal/validators"
)

// ProductService manages products and their images/colors.
type ProductService struct {
	products *repository.ProductRepository
	catalog  *repository.CatalogRepository
	files    storage.Storage // uploaded images (nil: uploads off)
	restock  Restocker       // waiting-list notifications (nil: off)
}

// Restocker is told when a product has stock again (the waiting list
// service implements it).
type Restocker interface {
	NotifyRestock(ctx context.Context, productID int64)
}

func NewProductService(products *repository.ProductRepository, catalog *repository.CatalogRepository,
	files storage.Storage, restock Restocker) *ProductService {
	return &ProductService{products: products, catalog: catalog, files: files, restock: restock}
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
	return models.ProductDetail{Product: p, Images: images}, nil
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
	// Back in stock: tell the customers on the waiting list.
	if in.Stock > 0 && (in.IsActive == nil || *in.IsActive) && s.restock != nil {
		s.restock.NotifyRestock(ctx, id)
	}
	return s.Get(ctx, id, true)
}

func (s *ProductService) Deactivate(ctx context.Context, id int64) error {
	return s.products.Deactivate(ctx, id)
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

// UploadImage checks an uploaded picture, stores it and adds it to the
// product's images. If saving the row fails, the stored file is removed again.
func (s *ProductService) UploadImage(ctx context.Context, productID int64, data []byte, isPrimary bool, displayOrder int) (models.ProductImage, error) {
	if displayOrder < 0 {
		return models.ProductImage{}, validators.FieldError("display_order", "must be 0 or more")
	}
	if err := s.requireProduct(ctx, productID, false); err != nil {
		return models.ProductImage{}, err
	}
	fileURL, err := saveImage(ctx, s.files, data)
	if err != nil {
		return models.ProductImage{}, err
	}
	img, err := s.catalog.AddImage(ctx, productID, models.ProductImageInput{
		ImageURL: fileURL, IsPrimary: isPrimary, DisplayOrder: displayOrder,
	})
	if err != nil {
		_ = s.files.Delete(context.WithoutCancel(ctx), fileURL)
		return models.ProductImage{}, err
	}
	return img, nil
}

// DeleteImage removes the image row and, if the file was uploaded to our
// storage, the file itself.
func (s *ProductService) DeleteImage(ctx context.Context, imageID int64) error {
	fileURL, err := s.catalog.DeleteImage(ctx, imageID)
	if err != nil {
		return err
	}
	if s.files != nil {
		if err := s.files.Delete(context.WithoutCancel(ctx), fileURL); err != nil {
			slog.Error("delete image file", "url", fileURL, "error", err) // the row is gone; a stray file is harmless
		}
	}
	return nil
}

// saveImage validates and stores an uploaded picture, returning its URL.
func saveImage(ctx context.Context, files storage.Storage, data []byte) (string, error) {
	if files == nil {
		return "", apperr.Conflict("image uploads are not configured")
	}
	img, err := images.Inspect(data)
	if err != nil {
		return "", validators.FieldError("image", err.Error())
	}
	return files.Save(ctx, img.Ext, img.ContentType, img.Data)
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
