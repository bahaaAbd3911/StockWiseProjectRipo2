package services

import (
	"context"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

type AddressService struct {
	addresses *repository.AddressRepository
}

func NewAddressService(addresses *repository.AddressRepository) *AddressService {
	return &AddressService{addresses: addresses}
}

func (s *AddressService) List(ctx context.Context, userID int64) ([]models.Address, error) {
	return s.addresses.List(ctx, userID)
}

func (s *AddressService) Get(ctx context.Context, userID, id int64) (models.Address, error) {
	return s.addresses.Get(ctx, userID, id)
}

func (s *AddressService) Create(ctx context.Context, userID int64, in models.AddressInput) (models.Address, error) {
	if err := validators.ValidateAddress(&in); err != nil {
		return models.Address{}, err
	}
	return s.addresses.Create(ctx, userID, in)
}

func (s *AddressService) Update(ctx context.Context, userID, id int64, in models.AddressInput) (models.Address, error) {
	if err := validators.ValidateAddress(&in); err != nil {
		return models.Address{}, err
	}
	return s.addresses.Update(ctx, userID, id, in)
}

func (s *AddressService) Delete(ctx context.Context, userID, id int64) error {
	return s.addresses.Delete(ctx, userID, id)
}

func (s *AddressService) SetDefault(ctx context.Context, userID, id int64) (models.Address, error) {
	return s.addresses.SetDefault(ctx, userID, id)
}
