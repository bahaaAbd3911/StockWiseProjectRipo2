package services

import (
	"context"
	"strings"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"

	"golang.org/x/crypto/bcrypt"
)

// AdminUserService is user management for admins. Two safety rules: an admin
// cannot demote, deactivate or delete their own account, and the last active
// admin can never be removed (the shop would have nobody to manage it).
type AdminUserService struct {
	users     *repository.AdminUserRepository
	orders    *repository.OrderRepository
	favorites *repository.FavoriteRepository
}

func NewAdminUserService(users *repository.AdminUserRepository, orders *repository.OrderRepository,
	favorites *repository.FavoriteRepository) *AdminUserService {
	return &AdminUserService{users: users, orders: orders, favorites: favorites}
}

func (s *AdminUserService) List(ctx context.Context, f models.AdminUserFilter) (models.Page[models.AdminUser], error) {
	f.Search = strings.TrimSpace(f.Search)
	f.Role = strings.ToLower(strings.TrimSpace(f.Role))
	if f.Role != "" && f.Role != auth.RoleAdmin && f.Role != auth.RoleCustomer {
		return models.Page[models.AdminUser]{}, validators.FieldError("role", "must be customer or admin")
	}
	return s.users.List(ctx, f)
}

func (s *AdminUserService) Details(ctx context.Context, id int64) (models.AdminUserDetails, error) {
	u, err := s.users.Get(ctx, id)
	if err != nil {
		return models.AdminUserDetails{}, err
	}
	orders, err := s.orders.ListForUser(ctx, id)
	if err != nil {
		return models.AdminUserDetails{}, err
	}
	favorites, err := s.favorites.List(ctx, id)
	if err != nil {
		return models.AdminUserDetails{}, err
	}
	return models.AdminUserDetails{User: u, Orders: orders, Favorites: favorites}, nil
}

func (s *AdminUserService) Create(ctx context.Context, in models.AdminUserInput) (models.AdminUser, error) {
	if err := validators.ValidateAdminUser(&in, true); err != nil {
		return models.AdminUser{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.AdminUser{}, err
	}
	id, err := s.users.Create(ctx, in, string(hash))
	if err != nil {
		return models.AdminUser{}, err
	}
	return s.users.Get(ctx, id)
}

func (s *AdminUserService) Update(ctx context.Context, adminID, id int64, in models.AdminUserInput) (models.AdminUser, error) {
	if err := validators.ValidateAdminUser(&in, false); err != nil {
		return models.AdminUser{}, err
	}
	current, err := s.users.Get(ctx, id)
	if err != nil {
		return models.AdminUser{}, err
	}
	losesAdmin := current.Role == auth.RoleAdmin && current.IsActive &&
		(in.Role != auth.RoleAdmin || (in.IsActive != nil && !*in.IsActive))
	if losesAdmin {
		if id == adminID {
			return models.AdminUser{}, apperr.Conflict("you cannot remove your own admin rights or deactivate your own account")
		}
		if err := s.keepOneAdmin(ctx, id); err != nil {
			return models.AdminUser{}, err
		}
	}
	var hash *string
	if in.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.AdminUser{}, err
		}
		hs := string(h)
		hash = &hs
	}
	if err := s.users.Update(ctx, id, in, hash); err != nil {
		return models.AdminUser{}, err
	}
	return s.users.Get(ctx, id)
}

// Delete removes the account, or deactivates it when it has orders. Returns
// the message to show.
func (s *AdminUserService) Delete(ctx context.Context, adminID, id int64) (string, error) {
	if id == adminID {
		return "", apperr.Conflict("you cannot delete your own account")
	}
	current, err := s.users.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if current.Role == auth.RoleAdmin && current.IsActive {
		if err := s.keepOneAdmin(ctx, id); err != nil {
			return "", err
		}
	}
	deleted, err := s.users.Delete(ctx, id)
	if err != nil {
		return "", err
	}
	if deleted {
		return "user deleted", nil
	}
	return "this user has orders, so the account was deactivated instead of deleted", nil
}

func (s *AdminUserService) keepOneAdmin(ctx context.Context, id int64) error {
	others, err := s.users.ActiveAdminsExcept(ctx, id)
	if err != nil {
		return err
	}
	if others == 0 {
		return apperr.Conflict("this is the last active admin; make another user admin first")
	}
	return nil
}
