package services

import (
	"errors"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
)

type AddressService struct {
	Repository *repository.AddressRepository
}

func NewAddressService(repo *repository.AddressRepository) *AddressService {
	return &AddressService{Repository: repo}
}

func (s *AddressService) CreateAddress(
	userID int64,
	request models.CreateAddressRequest,
) (models.Address, error) {
	if userID <= 0 {
		return models.Address{}, errors.New("invalid user ID")
	}

	if request.RecipientName == "" ||
		request.Phone == "" ||
		request.City == "" ||
		request.Area == "" ||
		request.Street == "" {
		return models.Address{}, errors.New("required address fields are missing")
	}

	return s.Repository.CreateAddress(userID, request)
}

func (s *AddressService) GetAddresses(userID int64) ([]models.Address, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.Repository.GetAddresses(userID)
}

func (s *AddressService) UpdateAddress(
	userID int64,
	addressID int64,
	request models.UpdateAddressRequest,
) (models.Address, error) {
	if userID <= 0 || addressID <= 0 {
		return models.Address{}, errors.New("invalid user ID or address ID")
	}

	if request.RecipientName == "" ||
		request.Phone == "" ||
		request.City == "" ||
		request.Area == "" ||
		request.Street == "" {
		return models.Address{}, errors.New("required address fields are missing")
	}

	return s.Repository.UpdateAddress(userID, addressID, request)
}

func (s *AddressService) DeleteAddress(userID int64, addressID int64) error {
	if userID <= 0 || addressID <= 0 {
		return errors.New("invalid user ID or address ID")
	}

	return s.Repository.DeleteAddress(userID, addressID)
}

func (s *AddressService) SetDefaultAddress(userID int64, addressID int64) error {
	if userID <= 0 || addressID <= 0 {
		return errors.New("invalid user ID or address ID")
	}

	return s.Repository.SetDefaultAddress(userID, addressID)
}
