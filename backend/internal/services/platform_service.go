package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/notify"
	"ecommerce-backend/internal/ratelimit"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/tenant"
	"ecommerce-backend/internal/validators"

	"golang.org/x/crypto/bcrypt"
)

// InviteTTL is how long a new store admin's "choose your password" link works.
const InviteTTL = 72 * time.Hour

// PlatformService is the platform super admin: log in, and create, edit,
// suspend and reactivate stores.
type PlatformService struct {
	repo          *repository.PlatformRepository
	tokens        *auth.TokenManager // platform tokens (a different issuer from store logins)
	mailer        mail.Mailer
	branding      *notify.Branding
	storeURL      func(tenant.Store) string
	resolver      *tenant.Resolver // forgotten after a store changes, so suspensions apply at once
	loginFailures *ratelimit.Limiter
}

func NewPlatformService(repo *repository.PlatformRepository, tokens *auth.TokenManager, mailer mail.Mailer,
	branding *notify.Branding, storeURL func(tenant.Store) string, resolver *tenant.Resolver,
	loginFailures *ratelimit.Limiter) *PlatformService {
	return &PlatformService{repo: repo, tokens: tokens, mailer: mailer, branding: branding,
		storeURL: storeURL, resolver: resolver, loginFailures: loginFailures}
}

var errPlatformCredentials = apperr.Unauthorized("invalid email or password")

// Login checks a platform admin's password (same protections as store
// logins: same timing for unknown emails, failures limited per email).
func (s *PlatformService) Login(ctx context.Context, req models.LoginRequest) (models.PlatformLoginResponse, error) {
	if err := validators.ValidateLogin(&req); err != nil {
		return models.PlatformLoginResponse{}, err
	}
	key := "platform:" + req.Email
	if ok, wait := s.loginFailures.Check(key); !ok {
		return models.PlatformLoginResponse{}, apperr.TooManyRequests("too many failed login attempts; please try again later", wait)
	}
	admin, hash, err := s.repo.AdminByEmail(ctx, req.Email)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(req.Password))
		s.loginFailures.Fail(key)
		return models.PlatformLoginResponse{}, errPlatformCredentials
	}
	if err != nil {
		return models.PlatformLoginResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		s.loginFailures.Fail(key)
		return models.PlatformLoginResponse{}, errPlatformCredentials
	}
	if !admin.IsActive {
		return models.PlatformLoginResponse{}, apperr.Unauthorized("this account has been deactivated")
	}
	s.loginFailures.Reset(key)
	token, expires, err := s.tokens.Issue(0, admin.ID, admin.TokenVersion)
	if err != nil {
		return models.PlatformLoginResponse{}, err
	}
	return models.PlatformLoginResponse{Admin: admin, Token: token, ExpiresAt: expires}, nil
}

func (s *PlatformService) Me(ctx context.Context, id int64) (models.PlatformAdmin, error) {
	a, err := s.repo.AdminByID(ctx, id)
	return a, apperr.FromDB(err, "admin not found")
}

// AuthInfo is used by the platform authentication middleware.
func (s *PlatformService) AuthInfo(ctx context.Context, id int64) (active bool, version int, err error) {
	a, err := s.repo.AdminByID(ctx, id)
	return a.IsActive, a.TokenVersion, err
}

func (s *PlatformService) withURL(st models.PlatformStore) models.PlatformStore {
	st.URL = s.storeURL(tenant.Store{ID: st.ID, Slug: st.Slug, PrimaryDomain: st.PrimaryDomain})
	return st
}

func (s *PlatformService) ListStores(ctx context.Context, f models.StoreFilter) (models.Page[models.PlatformStore], error) {
	f.Search = strings.TrimSpace(f.Search)
	f.Status = strings.TrimSpace(f.Status)
	if f.Status != "" && f.Status != tenant.StatusActive && f.Status != tenant.StatusSuspended {
		return models.Page[models.PlatformStore]{}, validators.FieldError("status", "must be active or suspended")
	}
	page, err := s.repo.ListStores(ctx, f)
	for i := range page.Items {
		page.Items[i] = s.withURL(page.Items[i])
	}
	return page, err
}

func (s *PlatformService) GetStore(ctx context.Context, id int64) (models.PlatformStore, error) {
	st, err := s.repo.GetStore(ctx, id)
	return s.withURL(st), err
}

// CreateStore creates a ready-to-use store: its record, settings and admin
// account, and emails the admin a link to choose a password. No new database,
// no new deployment: the store is live at its address right away.
func (s *PlatformService) CreateStore(ctx context.Context, in models.CreateStoreRequest) (models.CreatedStore, error) {
	if err := validators.ValidateCreateStore(&in); err != nil {
		return models.CreatedStore{}, err
	}
	token, tokenHash, err := newSecretToken()
	if err != nil {
		return models.CreatedStore{}, err
	}
	unusable, err := unusablePasswordHash()
	if err != nil {
		return models.CreatedStore{}, err
	}
	id, err := s.repo.CreateStore(ctx, in, repository.NewStore{
		AdminPasswordHash: unusable, InviteTokenHash: tokenHash, InviteExpires: time.Now().Add(InviteTTL),
	})
	if err != nil {
		return models.CreatedStore{}, err
	}
	st, err := s.GetStore(ctx, id)
	if err != nil {
		return models.CreatedStore{}, err
	}
	msg := "store created; an invitation was sent to " + in.Admin.Email
	if err := s.sendInvite(ctx, st, notify.Recipient{Email: in.Admin.Email, FirstName: in.Admin.FirstName}, token); err != nil {
		msg = "store created, but the invitation email could not be sent; use \"send the invite again\""
	}
	return models.CreatedStore{Store: st, Message: msg}, nil
}

// ResendInvite gives the store's first admin a new invite link (the old one
// stops working).
func (s *PlatformService) ResendInvite(ctx context.Context, id int64) (string, error) {
	st, err := s.GetStore(ctx, id)
	if err != nil {
		return "", err
	}
	token, tokenHash, err := newSecretToken()
	if err != nil {
		return "", err
	}
	admin, err := s.repo.NewInvite(ctx, id, tokenHash, time.Now().Add(InviteTTL))
	if err != nil {
		return "", err
	}
	if err := s.sendInvite(ctx, st, notify.Recipient{Email: admin.Email, FirstName: admin.FirstName}, token); err != nil {
		return "", err
	}
	return "a new invitation was sent to " + admin.Email, nil
}

// The invite carries a secret link, so it is sent directly (not queued, like
// password resets).
func (s *PlatformService) sendInvite(ctx context.Context, st models.PlatformStore, to notify.Recipient, token string) error {
	tpl := s.branding.ForStore(tenant.Store{ID: st.ID, Slug: st.Slug, Name: st.Name, Currency: st.Currency, PrimaryDomain: st.PrimaryDomain})
	link, err := notify.AddQuery(tpl.ResetURL, "token", token)
	if err != nil {
		return err
	}
	return s.mailer.Send(ctx, tpl.StoreInvite(to, link, int(InviteTTL.Hours())))
}

func (s *PlatformService) UpdateStore(ctx context.Context, id int64, in models.UpdateStoreRequest) (models.PlatformStore, error) {
	if err := validators.ValidateUpdateStore(&in); err != nil {
		return models.PlatformStore{}, err
	}
	if err := s.repo.UpdateStore(ctx, id, in); err != nil {
		return models.PlatformStore{}, err
	}
	s.resolver.Forget()
	return s.GetStore(ctx, id)
}

// SetStatus suspends or reactivates a store. Suspended stores answer 403 to
// everyone; nothing is deleted.
func (s *PlatformService) SetStatus(ctx context.Context, id int64, status string) (models.PlatformStore, error) {
	if err := s.repo.SetStoreStatus(ctx, id, status); err != nil {
		return models.PlatformStore{}, err
	}
	s.resolver.Forget()
	return s.GetStore(ctx, id)
}

// unusablePasswordHash: a bcrypt hash of random bytes nobody knows, so the
// account cannot be logged into until the admin sets a password.
func unusablePasswordHash() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	h, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(b)[:64]), bcrypt.DefaultCost)
	return string(h), err
}
