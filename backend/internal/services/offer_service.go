package services

import (
	"context"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// OfferService manages timed promotions. Offers are applied automatically by
// the product, cart and checkout code (see pricing.BestPrice); this service is
// only about creating, editing and listing them.
type OfferService struct {
	offers   *repository.OfferRepository
	products *repository.ProductRepository
}

func NewOfferService(offers *repository.OfferRepository, products *repository.ProductRepository) *OfferService {
	return &OfferService{offers: offers, products: products}
}

// ListLive is the public list of offers running right now.
func (s *OfferService) ListLive(ctx context.Context, limit, offset int) (models.Page[models.Offer], error) {
	return s.offers.List(ctx, models.OfferFilter{Status: models.OfferLive, Limit: limit, Offset: offset})
}

// GetLive returns a running offer and a page of the active products it covers.
func (s *OfferService) GetLive(ctx context.Context, id int64, limit, offset int) (models.Offer, models.Page[models.Product], error) {
	o, err := s.offers.Get(ctx, id)
	if err != nil {
		return o, models.Page[models.Product]{}, err
	}
	if o.Status != models.OfferLive {
		return o, models.Page[models.Product]{}, apperr.NotFound("offer not found")
	}
	products, err := s.products.List(ctx, models.ProductFilter{OfferID: o.ID, Sort: "newest", Limit: limit, Offset: offset})
	return o, products, err
}

func (s *OfferService) List(ctx context.Context, f models.OfferFilter) (models.Page[models.Offer], error) {
	f.Status = strings.TrimSpace(f.Status)
	if f.Status != "" && !repository.IsValidOfferStatus(f.Status) {
		return models.Page[models.Offer]{}, validators.FieldError("status", "must be one of: live, scheduled, expired, disabled")
	}
	return s.offers.List(ctx, f)
}

func (s *OfferService) Get(ctx context.Context, id int64) (models.Offer, error) {
	return s.offers.Get(ctx, id)
}

func (s *OfferService) Create(ctx context.Context, in models.OfferInput) (models.Offer, error) {
	if err := validators.ValidateOffer(&in); err != nil {
		return models.Offer{}, err
	}
	id, err := s.offers.Create(ctx, in)
	if err != nil {
		return models.Offer{}, err
	}
	return s.offers.Get(ctx, id)
}

func (s *OfferService) Update(ctx context.Context, id int64, in models.OfferInput) (models.Offer, error) {
	if err := validators.ValidateOffer(&in); err != nil {
		return models.Offer{}, err
	}
	if err := s.offers.Update(ctx, id, in); err != nil {
		return models.Offer{}, err
	}
	return s.offers.Get(ctx, id)
}

func (s *OfferService) Delete(ctx context.Context, id int64) error {
	return s.offers.Delete(ctx, id)
}
