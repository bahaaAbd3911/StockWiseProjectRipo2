package services

import (
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/utils"

	"errors"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repository *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		Repository: repo,
	}
}

func (s *UserService) RegisterUser(
	user models.RegisterUserRequest,
) (models.User, error) {

	// Hash the user's password
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return models.User{}, err
	}

	// Save the user with the hashed password
	return s.Repository.CreateUser(
		user,
		string(passwordHash),
	)
}
func (s *UserService) LoginUser(login models.LoginRequest) (models.LoginResponse, error) {
	user, passwordHash, err := s.Repository.GetUserByEmail(login.Email)
	if err != nil {
		return models.LoginResponse{}, err
	}

	if !user.IsActive {
		return models.LoginResponse{}, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(login.Password),
	)
	if err != nil {
		return models.LoginResponse{}, errors.New("invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}
func (s *UserService) GetUserProfile(userID int64) (models.User, error) {
	return s.Repository.GetUserByID(userID)
}
func (s *UserService) UpdateUserProfile(
	userID int64,
	request models.UpdateProfileRequest,
) (models.User, error) {
	return s.Repository.UpdateUserProfile(userID, request)
}
func (s *UserService) ChangePassword(
	userID int64,
	request models.ChangePasswordRequest,
) error {
	currentHash, err := s.Repository.GetPasswordHashByID(userID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(currentHash),
		[]byte(request.CurrentPassword),
	)
	if err != nil {
		return errors.New("current password is incorrect")
	}

	if len(request.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	newHash, err := bcrypt.GenerateFromPassword(
		[]byte(request.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	return s.Repository.UpdatePassword(userID, string(newHash))
}
func (s *UserService) DeactivateUser(userID int64) error {
	return s.Repository.DeactivateUser(userID)
}
