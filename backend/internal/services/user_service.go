// Package services holds the business rules. A service validates input,
// decides what should happen, and calls one or more repositories. It knows
// nothing about HTTP (no status codes, no requests) and nothing about SQL.
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/notify"
	"ecommerce-backend/internal/ratelimit"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/storage"
	"ecommerce-backend/internal/tenant"
	"ecommerce-backend/internal/validators"

	"golang.org/x/crypto/bcrypt"
)

// UserServiceOptions are the account-security settings.
type UserServiceOptions struct {
	Mailer mail.Mailer
	// Branding: each store's name and the pages its emailed links open
	// (reset password, confirm email) on that store's own address.
	Branding *notify.Branding

	ResetTTL  time.Duration // how long a reset link works
	VerifyTTL time.Duration // how long a confirmation link works

	LoginFailures *ratelimit.Limiter // per email
	ResetRequests *ratelimit.Limiter // per email
	VerifyResends *ratelimit.Limiter // per user

	Files storage.Storage // uploaded profile pictures (nil: uploads off)
}

type UserService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
	opts   UserServiceOptions
	// background tracks emails being prepared and sent (shutdown and tests wait on it).
	background sync.WaitGroup
}

func NewUserService(users *repository.UserRepository, tokens *auth.TokenManager, opts UserServiceOptions) *UserService {
	if opts.Mailer == nil {
		opts.Mailer = mail.LogMailer{}
	}
	if opts.ResetTTL <= 0 {
		opts.ResetTTL = 30 * time.Minute
	}
	if opts.VerifyTTL <= 0 {
		opts.VerifyTTL = 48 * time.Hour
	}
	return &UserService{users: users, tokens: tokens, opts: opts}
}

var errBadCredentials = apperr.Unauthorized("invalid email or password")

// dummyHash is compared against when the email does not exist, so a login for
// an unknown email takes as long as one with a wrong password. Otherwise the
// response time would reveal which emails have accounts.
var dummyHash = sync.OnceValue(func() []byte {
	h, _ := bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)
	return h
})

// Register creates the account and emails a confirmation link. The account
// works right away (login, cart, favorites); placing an order needs the
// confirmed email when REQUIRE_EMAIL_VERIFICATION is on.
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
	user, err := s.users.Create(ctx, req, string(hash))
	if err != nil {
		return models.User{}, err
	}
	s.inBackground(ctx, "confirmation email", func(bg context.Context) error { return s.sendVerification(bg, user) })
	return user, nil
}

// Login checks the password and issues a token. Failed attempts are counted
// per email (on top of the per-IP limit): after LoginFailures attempts the
// account refuses logins for a while, which stops password guessing spread
// over many IPs. A successful login clears the counter.
func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	if err := validators.ValidateLogin(&req); err != nil {
		return models.LoginResponse{}, err
	}
	if ok, wait := s.opts.LoginFailures.Check(perStore(ctx, req.Email)); !ok {
		return models.LoginResponse{}, apperr.TooManyRequests(
			"too many failed login attempts for this account; please try again later or reset your password", wait)
	}

	user, hash, err := s.users.GetByEmail(ctx, req.Email)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(req.Password))
		s.opts.LoginFailures.Fail(perStore(ctx, req.Email))
		return models.LoginResponse{}, errBadCredentials
	}
	if err != nil {
		return models.LoginResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		s.opts.LoginFailures.Fail(perStore(ctx, req.Email))
		return models.LoginResponse{}, errBadCredentials
	}
	if !user.IsActive {
		return models.LoginResponse{}, apperr.Unauthorized("this account has been deactivated")
	}
	s.opts.LoginFailures.Reset(perStore(ctx, req.Email))

	token, expires, err := s.tokens.Issue(tenant.ID(ctx), user.ID, user.TokenVersion)
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

// ChangePassword sets a new password and logs the user out on every OTHER
// device: all existing tokens are revoked, and a fresh token is returned so
// the device that made the change stays logged in.
func (s *UserService) ChangePassword(ctx context.Context, userID int64, req models.ChangePasswordRequest) (models.TokenResponse, error) {
	if err := validators.ValidateChangePassword(&req); err != nil {
		return models.TokenResponse{}, err
	}
	current, err := s.users.PasswordHash(ctx, userID)
	if err != nil {
		return models.TokenResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(current), []byte(req.CurrentPassword)) != nil {
		// 400, not 401: a 401 would make most frontends log the user out.
		return models.TokenResponse{}, apperr.Validation(map[string]string{"current_password": "is incorrect"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return models.TokenResponse{}, err
	}
	version, err := s.users.UpdatePassword(ctx, userID, string(hash))
	if err != nil {
		return models.TokenResponse{}, err
	}
	token, expires, err := s.tokens.Issue(tenant.ID(ctx), userID, version)
	if err != nil {
		return models.TokenResponse{}, err
	}
	return models.TokenResponse{
		Message: "password changed; you have been logged out on all other devices",
		Token:   token, ExpiresAt: expires,
	}, nil
}

// LogoutEverywhere revokes every token of the user, including the current one.
func (s *UserService) LogoutEverywhere(ctx context.Context, userID int64) error {
	return s.users.RevokeTokens(ctx, userID)
}

// UploadProfileImage replaces the profile picture with an uploaded one. The
// previous picture's file is removed if it was uploaded too.
func (s *UserService) UploadProfileImage(ctx context.Context, userID int64, data []byte) (models.User, error) {
	fileURL, err := saveImage(ctx, s.opts.Files, data)
	if err != nil {
		return models.User{}, err
	}
	old, err := s.users.SetProfileImage(ctx, userID, &fileURL)
	if err != nil {
		_ = s.opts.Files.Delete(context.WithoutCancel(ctx), fileURL)
		return models.User{}, err
	}
	s.deleteFile(ctx, old)
	return s.users.GetByID(ctx, userID)
}

// DeleteProfileImage removes the profile picture.
func (s *UserService) DeleteProfileImage(ctx context.Context, userID int64) (models.User, error) {
	old, err := s.users.SetProfileImage(ctx, userID, nil)
	if err != nil {
		return models.User{}, err
	}
	s.deleteFile(ctx, old)
	return s.users.GetByID(ctx, userID)
}

func (s *UserService) deleteFile(ctx context.Context, fileURL *string) {
	if fileURL == nil || s.opts.Files == nil {
		return
	}
	if err := s.opts.Files.Delete(context.WithoutCancel(ctx), *fileURL); err != nil {
		slog.Error("delete old profile image", "error", err)
	}
}

func (s *UserService) Deactivate(ctx context.Context, userID int64) error {
	return s.users.Deactivate(ctx, userID)
}

// ---------- email confirmation ----------

// VerifyEmail confirms the address with the token from the emailed link.
func (s *UserService) VerifyEmail(ctx context.Context, req models.VerifyEmailRequest) error {
	if req.Token == "" {
		return validators.FieldError("token", "is required")
	}
	return s.users.VerifyEmail(ctx, hashToken(req.Token))
}

// ResendVerification emails a new confirmation link (the old one stops
// working). Limited per user, so it cannot be used to flood a mailbox.
func (s *UserService) ResendVerification(ctx context.Context, userID int64) (string, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user.EmailVerified {
		return "your email address is already confirmed", nil
	}
	if ok, wait := s.opts.VerifyResends.Allow(strconv.FormatInt(userID, 10)); !ok {
		return "", apperr.TooManyRequests("a confirmation email was sent recently; please check your inbox (and spam folder)", wait)
	}
	s.inBackground(ctx, "confirmation email", func(bg context.Context) error { return s.sendVerification(bg, user) })
	return "a new confirmation link has been sent to " + user.Email, nil
}

func (s *UserService) sendVerification(ctx context.Context, user models.User) error {
	token, hash, err := newSecretToken()
	if err != nil {
		return err
	}
	if err := s.users.CreateVerifyToken(ctx, user.ID, hash, time.Now().Add(s.opts.VerifyTTL)); err != nil {
		return err
	}
	tpl := s.opts.Branding.For(ctx)
	link, err := notify.AddQuery(tpl.VerifyURL, "token", token)
	if err != nil {
		return err
	}
	to := notify.Recipient{Email: user.Email, FirstName: user.FirstName}
	return s.opts.Mailer.Send(ctx, tpl.VerifyEmail(to, link, int(s.opts.VerifyTTL.Hours())))
}

// ---------- password reset ----------

// ForgotPasswordMessage is returned for EVERY valid request, whether or not the
// email has an account, so the endpoint cannot be used to find out who is a
// customer.
const ForgotPasswordMessage = "if an account exists for this email, a password reset link has been sent"

// RequestPasswordReset starts a reset. The work (lookup, token, email) runs in
// the background so the response time is the same for known and unknown
// emails. Requests per email are limited (silently: the reply never changes).
func (s *UserService) RequestPasswordReset(ctx context.Context, req models.ForgotPasswordRequest, ip string) error {
	if err := validators.ValidateForgotPassword(&req); err != nil {
		return err
	}
	if ok, _ := s.opts.ResetRequests.Allow(perStore(ctx, req.Email)); !ok {
		return nil
	}
	s.inBackground(ctx, "password reset email", func(bg context.Context) error { return s.sendResetLink(bg, req.Email, ip) })
	return nil
}

func (s *UserService) sendResetLink(ctx context.Context, email, ip string) error {
	user, _, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // no account: nothing to send, nothing to reveal
	}
	if err != nil {
		return err
	}
	if !user.IsActive {
		return nil
	}

	token, hash, err := newSecretToken()
	if err != nil {
		return err
	}
	if err := s.users.CreateResetToken(ctx, user.ID, hash, time.Now().Add(s.opts.ResetTTL), ip); err != nil {
		return err
	}
	tpl := s.opts.Branding.For(ctx)
	link, err := notify.AddQuery(tpl.ResetURL, "token", token)
	if err != nil {
		return err
	}
	to := notify.Recipient{Email: user.Email, FirstName: user.FirstName}
	return s.opts.Mailer.Send(ctx, tpl.PasswordReset(to, link, int(s.opts.ResetTTL.Minutes())))
}

// ResetPassword finishes a reset with the token from the email link. On
// success the user is logged out everywhere and must log in with the new
// password.
func (s *UserService) ResetPassword(ctx context.Context, req models.ResetPasswordRequest) error {
	if err := validators.ValidateResetPassword(&req); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.ResetPassword(ctx, hashToken(req.Token), string(hash))
}

// ---------- helpers ----------

// perStore makes a rate-limit key per store: the same email in two stores is
// two different accounts.
func perStore(ctx context.Context, key string) string {
	return strconv.FormatInt(tenant.ID(ctx), 10) + ":" + key
}

// inBackground runs fn after the response has been sent. It gets its own
// 30-second context, because the request's context ends with the response.
func (s *UserService) inBackground(ctx context.Context, what string, fn func(context.Context) error) {
	s.background.Add(1)
	go func() {
		defer s.background.Done()
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		if err := fn(bg); err != nil {
			slog.Error(what+" failed", "error", err)
		}
	}()
}

// WaitForBackground blocks until emails being sent are done (used by
// graceful shutdown and tests).
func (s *UserService) WaitForBackground() { s.background.Wait() }

// newSecretToken returns a random token for an emailed link and the SHA-256
// hash that is stored. 32 random bytes cannot be guessed.
func newSecretToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
