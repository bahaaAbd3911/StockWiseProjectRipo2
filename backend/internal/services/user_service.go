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
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"ecommerce-backend/internal/apperr"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/ratelimit"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"

	"golang.org/x/crypto/bcrypt"
)

// UserServiceOptions are the account-security settings.
type UserServiceOptions struct {
	Mailer        mail.Mailer
	ResetURL      string        // frontend page that receives ?token=...
	ResetTTL      time.Duration // how long a reset link works
	LoginFailures *ratelimit.Limiter
	ResetRequests *ratelimit.Limiter
}

type UserService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
	opts   UserServiceOptions
	// background tracks reset emails being sent (tests and shutdown wait on it).
	background sync.WaitGroup
}

func NewUserService(users *repository.UserRepository, tokens *auth.TokenManager, opts UserServiceOptions) *UserService {
	if opts.Mailer == nil {
		opts.Mailer = mail.LogMailer{}
	}
	if opts.ResetTTL <= 0 {
		opts.ResetTTL = 30 * time.Minute
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

// Login checks the password and issues a token. Failed attempts are counted
// per email (on top of the per-IP limit): after LoginFailures attempts the
// account refuses logins for a while, which stops password guessing spread
// over many IPs. A successful login clears the counter.
func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	if err := validators.ValidateLogin(&req); err != nil {
		return models.LoginResponse{}, err
	}
	if ok, wait := s.opts.LoginFailures.Check(req.Email); !ok {
		return models.LoginResponse{}, apperr.TooManyRequests(
			"too many failed login attempts for this account; please try again later or reset your password", wait)
	}

	user, hash, err := s.users.GetByEmail(ctx, req.Email)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(req.Password))
		s.opts.LoginFailures.Fail(req.Email)
		return models.LoginResponse{}, errBadCredentials
	}
	if err != nil {
		return models.LoginResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		s.opts.LoginFailures.Fail(req.Email)
		return models.LoginResponse{}, errBadCredentials
	}
	if !user.IsActive {
		return models.LoginResponse{}, apperr.Unauthorized("this account has been deactivated")
	}
	s.opts.LoginFailures.Reset(req.Email)

	token, expires, err := s.tokens.Issue(user.ID, user.TokenVersion)
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
	token, expires, err := s.tokens.Issue(userID, version)
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

func (s *UserService) Deactivate(ctx context.Context, userID int64) error {
	return s.users.Deactivate(ctx, userID)
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
	if ok, _ := s.opts.ResetRequests.Allow(req.Email); !ok {
		return nil
	}
	s.background.Add(1)
	go func() {
		defer s.background.Done()
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		if err := s.sendResetLink(bg, req.Email, ip); err != nil {
			slog.Error("password reset email failed", "error", err)
		}
	}()
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

	token, hash, err := newResetToken()
	if err != nil {
		return err
	}
	expires := time.Now().Add(s.opts.ResetTTL)
	if err := s.users.CreateResetToken(ctx, user.ID, hash, expires, ip); err != nil {
		return err
	}
	link, err := resetLink(s.opts.ResetURL, token)
	if err != nil {
		return err
	}
	minutes := int(s.opts.ResetTTL.Minutes())
	return s.opts.Mailer.Send(ctx, mail.Message{
		To:      user.Email,
		Subject: "Reset your password / إعادة تعيين كلمة المرور",
		Text: fmt.Sprintf(`Hello %s,

Someone asked to reset the password of your account. If it was you, open this
link within %d minutes:

%s

If you did not ask for this, ignore this email - your password stays the same.

---

مرحباً %s،

طلب أحدهم إعادة تعيين كلمة المرور لحسابك. إذا كنت أنت، افتح الرابط التالي خلال %d دقيقة:

%s

إذا لم تطلب ذلك، تجاهل هذه الرسالة - لن تتغير كلمة المرور.
`, user.FirstName, minutes, link, user.FirstName, minutes, link),
	})
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

// WaitForBackground blocks until queued reset emails are handled (used by
// graceful shutdown and tests).
func (s *UserService) WaitForBackground() { s.background.Wait() }

// newResetToken returns a random token for the email link and the SHA-256
// hash that is stored. 32 random bytes cannot be guessed.
func newResetToken() (token, hash string, err error) {
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

func resetLink(base, token string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
