// Package services holds the business rules. A service validates input,
// decides what should happen, and calls one or more repositories. It knows
// nothing about HTTP (no status codes, no requests) and nothing about SQL.
package services

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
}

func NewUserService(users *repository.UserRepository, tokens *auth.TokenManager) *UserService {
	return &UserService{users: users, tokens: tokens}
}

var errBadCredentials = apperr.Unauthorized("invalid email or password")

// dummyHash is compared against when the email does not exist, so a login for
// an unknown email takes as long as one with a wrong password. Otherwise the
// response time would reveal which emails have accounts.
var dummyHash = sync.OnceValue(func() []byte {
	h, _ := bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)
	return h
})

func (s *UserService) Register(ctx context.Context, req models.RegisterRequest) (models.User, error) {
	if err := validators.ValidateRegister(&req); err != nil {
		return models.User{}, err
	}
	taken, err := s.users.EmailTaken(ctx, req.Email)
	if err != nil {
		return models.User{}, err
	}
	if taken {
		return models.User{}, apperr.Conflict("an account with this email already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	return s.users.Create(ctx, req, string(hash))
}

func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	if err := validators.ValidateLogin(&req); err != nil {
		return models.LoginResponse{}, err
	}
	user, hash, err := s.users.GetByEmail(ctx, req.Email)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(req.Password))
		return models.LoginResponse{}, errBadCredentials
	}
	if err != nil {
		return models.LoginResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		return models.LoginResponse{}, errBadCredentials
	}
	if !user.IsActive {
		return models.LoginResponse{}, apperr.Unauthorized("this account has been deactivated")
	}
	token, expires, err := s.tokens.Issue(user.ID)
	if err != nil {
		return models.LoginResponse{}, err
	}
	return models.LoginResponse{User: user, Token: token, ExpiresAt: expires}, nil
}

func (s *UserService) Profile(ctx context.Context, userID int64) (models.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req models.UpdateProfileRequest) (models.User, error) {
	if err := validators.ValidateUpdateProfile(&req); err != nil {
		return models.User{}, err
	}
	return s.users.UpdateProfile(ctx, userID, req)
}

func (s *UserService) ChangePassword(ctx context.Context, userID int64, req models.ChangePasswordRequest) error {
	if err := validators.ValidateChangePassword(&req); err != nil {
		return err
	}
	current, err := s.users.PasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(current), []byte(req.CurrentPassword)) != nil {
		// 400, not 401: a 401 would make most frontends log the user out.
		return apperr.Validation(map[string]string{"current_password": "is incorrect"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, string(hash))
}

func (s *UserService) Deactivate(ctx context.Context, userID int64) error {
	return s.users.Deactivate(ctx, userID)
}
