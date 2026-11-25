package services

import (
	"context"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// CategoryService: the public category list and admin category management.
type CategoryService struct {
	categories *repository.CategoryRepository
}

func NewCategoryService(categories *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

// List: public callers see active categories; admins see all of them.
func (s *CategoryService) List(ctx context.Context, includeInactive bool) ([]models.Category, error) {
	return s.categories.List(ctx, includeInactive)
}

func (s *CategoryService) Create(ctx context.Context, in models.CategoryInput) (models.Category, error) {
	if err := validators.ValidateCategory(&in); err != nil {
		return models.Category{}, err
	}
	id, err := s.categories.Create(ctx, in)
	if err != nil {
		return models.Category{}, err
	}
	return s.categories.Get(ctx, id)
}

// Update edits a category; renaming it also renames it on its products.
func (s *CategoryService) Update(ctx context.Context, id int64, in models.CategoryInput) (models.Category, error) {
	if err := validators.ValidateCategory(&in); err != nil {
		return models.Category{}, err
	}
	if err := s.categories.Update(ctx, id, in); err != nil {
		return models.Category{}, err
	}
	return s.categories.Get(ctx, id)
}

// Delete removes the category and hides its products (they are kept for
// order history and can be moved and re-activated).
func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	return s.categories.Delete(ctx, id)
}
