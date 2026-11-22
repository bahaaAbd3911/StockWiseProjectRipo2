package services

import (
	"context"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// BundleService manages bundles (several products sold together at one price).
// Putting bundles in the cart lives in CartService.
type BundleService struct {
	bundles *repository.BundleRepository
}

func NewBundleService(bundles *repository.BundleRepository) *BundleService {
	return &BundleService{bundles: bundles}
}

func (s *BundleService) List(ctx context.Context, f models.BundleFilter) (models.Page[models.Bundle], error) {
	return s.bundles.List(ctx, f)
}

func (s *BundleService) Get(ctx context.Context, id int64, includeUnavailable bool) (models.Bundle, error) {
	return s.bundles.Get(ctx, id, includeUnavailable)
}

func (s *BundleService) Create(ctx context.Context, in models.BundleInput) (models.Bundle, error) {
	if err := validators.ValidateBundle(&in); err != nil {
		return models.Bundle{}, err
	}
	id, err := s.bundles.Create(ctx, in)
	if err != nil {
		return models.Bundle{}, err
	}
	return s.bundles.Get(ctx, id, true)
}

func (s *BundleService) Update(ctx context.Context, id int64, in models.BundleInput) (models.Bundle, error) {
	if err := validators.ValidateBundle(&in); err != nil {
		return models.Bundle{}, err
	}
	if err := s.bundles.Update(ctx, id, in); err != nil {
		return models.Bundle{}, err
	}
	return s.bundles.Get(ctx, id, true)
}

func (s *BundleService) Deactivate(ctx context.Context, id int64) error {
	return s.bundles.Deactivate(ctx, id)
}
