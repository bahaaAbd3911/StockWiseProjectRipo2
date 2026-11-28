// Package routes builds the whole application: it creates every repository,
// service and handler (dependency injection), maps URLs to handlers, and wraps
// everything in the global middleware. Reading this file top to bottom shows
// every endpoint the API offers and who may call it.
package routes

import (
	"database/sql"
	"net/http"
	"net/url"
	"strings"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/handlers"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/mail"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/notify"
	"ecommerce-backend/internal/ratelimit"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/services"
	"ecommerce-backend/internal/storage"
	"ecommerce-backend/internal/tenant"
)

// StoreURL is a store's public frontend address: the default store (and any
// store when no platform is configured) uses FRONTEND_URL; others STORE_URL
// with their slug, e.g. https://ahmad.myplatform.com.
// A store with a verified main custom domain uses that domain.
func StoreURL(cfg config.Config, s tenant.Store) string {
	if s.PrimaryDomain != "" {
		scheme := "https"
		if strings.HasPrefix(cfg.FrontendURL, "http://") {
			scheme = "http"
		}
		return scheme + "://" + s.PrimaryDomain
	}
	if cfg.StoreURL == "" || s.Slug == "" || s.Slug == cfg.DefaultStore {
		return cfg.FrontendURL
	}
	return strings.ReplaceAll(cfg.StoreURL, "{slug}", s.Slug)
}

// StoreHost is a store's host name on the platform (what a custom domain's
// CNAME record points to), e.g. ahmad.myplatform.com.
func StoreHost(cfg config.Config, s tenant.Store) string {
	if cfg.PlatformDomain != "" && s.Slug != "" {
		return s.Slug + "." + cfg.PlatformDomain
	}
	if u, err := url.Parse(cfg.FrontendURL); err == nil {
		return u.Hostname()
	}
	return ""
}

// App is the assembled application.
type App struct {
	Handler http.Handler
	// Notifications delivers queued emails in the background (started by New).
	Notifications *notify.Worker
	users         *services.UserService
}

// Shutdown stops background work: it waits for account emails being sent and
// stops the notification worker (unsent emails stay queued for next start).
func (a *App) Shutdown() {
	a.users.WaitForBackground()
	a.Notifications.Stop()
}

// New builds the whole application. mailer delivers every email.
func New(db *sql.DB, cfg config.Config, mailer mail.Mailer) (*App, error) {
	// ---- rate limiters (nil = unlimited, used when limits are disabled) ----
	var general, authLimit, loginFailures, resetRequests, verifyResends *ratelimit.Limiter
	if rl := cfg.RateLimits; rl.Enabled {
		general = ratelimit.New(rl.General)             // per IP, every request
		authLimit = ratelimit.New(rl.Auth)              // per IP, login/register/password
		loginFailures = ratelimit.New(rl.LoginFailures) // per email, failed logins
		resetRequests = ratelimit.New(rl.ResetRequests) // per email, reset emails
		verifyResends = ratelimit.New(rl.ResetRequests) // per user, "send the confirmation email again"
	}

	// ---- files and emails ----
	files, err := storage.New(cfg.Storage)
	if err != nil {
		return nil, err
	}
	// Emails carry the name, address and currency of the store they are for.
	tpl := notify.NewBranding(
		notify.Templates{Store: cfg.StoreName, FrontendURL: cfg.FrontendURL, Currency: cfg.Currency},
		func(s tenant.Store) string { return StoreURL(cfg, s) },
		cfg.EmailVerifyURL, cfg.PasswordResetURL)
	outbox := notify.NewOutbox(db)
	worker := notify.NewWorker(db, mailer, notify.WorkerOptions{
		Interval:          cfg.EmailPollInterval,
		DebtReminderEvery: cfg.DebtReminderEvery,
		Branding:          tpl,
	})

	// Finds the store of each request from its host name (middleware.Store).
	resolver := tenant.NewResolver(db, cfg.PlatformDomain, cfg.DefaultStore)

	// ---- repositories (SQL) ----
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	catalogRepo := repository.NewCatalogRepository(db)
	cartRepo := repository.NewCartRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	waitingRepo := repository.NewWaitingListRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	bundleRepo := repository.NewBundleRepository(db)
	debtRepo := repository.NewDebtRepository(db)

	// ---- services (business rules) ----
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	userSvc := services.NewUserService(userRepo, tokens, services.UserServiceOptions{
		Mailer:        mailer,
		Branding:      tpl,
		ResetTTL:      cfg.PasswordResetTTL,
		VerifyTTL:     cfg.EmailVerifyTTL,
		LoginFailures: loginFailures,
		ResetRequests: resetRequests,
		VerifyResends: verifyResends,
		Files:         files,
	})
	waitingSvc := services.NewWaitingListService(waitingRepo, userRepo, outbox, tpl)
	productSvc := services.NewProductService(productRepo, catalogRepo, files, waitingSvc)
	cartSvc := services.NewCartService(cartRepo)
	favoriteSvc := services.NewFavoriteService(favoriteRepo)
	addressSvc := services.NewAddressService(addressRepo)
	orderSvc := services.NewOrderService(orderRepo, userRepo, outbox, tpl)
	reviewSvc := services.NewReviewService(reviewRepo, catalogRepo)
	offerSvc := services.NewOfferService(offerRepo, productRepo)
	bundleSvc := services.NewBundleService(bundleRepo)
	debtSvc := services.NewDebtService(debtRepo, userRepo, outbox, tpl)

	// ---- handlers (HTTP) ----
	health := handlers.NewHealthHandler(db)
	users := handlers.NewUserHandler(userSvc)
	products := handlers.NewProductHandler(productSvc)
	carts := handlers.NewCartHandler(cartSvc)
	favorites := handlers.NewFavoriteHandler(favoriteSvc)
	addresses := handlers.NewAddressHandler(addressSvc)
	orders := handlers.NewOrderHandler(orderSvc)
	reviews := handlers.NewReviewHandler(reviewSvc)
	waiting := handlers.NewWaitingListHandler(waitingSvc)
	offers := handlers.NewOfferHandler(offerSvc)
	bundles := handlers.NewBundleHandler(bundleSvc)
	debts := handlers.NewDebtHandler(debtSvc)
	settingsSvc := services.NewSettingsService(repository.NewSettingsRepository(db), files)
	uploads := handlers.NewUploadHandler(productSvc, userSvc, settingsSvc, cfg.MaxUploadBytes)
	emails := handlers.NewEmailHandler(services.NewEmailService(outbox))
	categories := handlers.NewCategoryHandler(services.NewCategoryService(repository.NewCategoryRepository(db)))
	settings := handlers.NewSettingsHandler(settingsSvc)
	adminUsers := handlers.NewAdminUserHandler(services.NewAdminUserService(
		repository.NewAdminUserRepository(db), orderRepo, favoriteRepo))
	coupons := handlers.NewCouponHandler(services.NewCouponService(repository.NewCouponRepository(db), cartRepo))
	dashboard := handlers.NewDashboardHandler(services.NewDashboardService(
		repository.NewDashboardRepository(db), cfg.StoreTimezone, cfg.LowStockThreshold))

	// ---- access levels ----
	mux := http.NewServeMux()
	authenticate := middleware.Authenticate(tokens, userRepo)

	public := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, httpx.Handle(h))
	}
	// sensitive: public, plus the strict per-IP limit for account endpoints.
	sensitive := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, middleware.RateLimit(authLimit)(httpx.Handle(h)))
	}
	user := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, authenticate(httpx.Handle(h)))
	}
	admin := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, authenticate(middleware.RequireAdmin(httpx.Handle(h))))
	}
	// userLimited: logged in, plus the strict per-IP limit (stops guessing coupon codes).
	userLimited := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, middleware.RateLimit(authLimit)(authenticate(httpx.Handle(h))))
	}
	// verified: logged in AND email confirmed (when REQUIRE_EMAIL_VERIFICATION is on).
	verified := user
	if cfg.RequireEmailVerification {
		verified = func(pattern string, h httpx.HandlerFunc) {
			mux.Handle(pattern, authenticate(middleware.RequireVerifiedEmail(httpx.Handle(h))))
		}
	}

	// Health
	public("GET /{$}", health.Root)
	public("GET /api/health", health.Health)

	// Uploaded files (local storage only; with S3 the bucket serves them)
	if local, ok := files.(*storage.Local); ok {
		mux.Handle("GET /uploads/{store}/{name}", local.Handler()) // per-store folders
		mux.Handle("GET /uploads/{name}", local.Handler())         // files from before multi-store
	}

	// Account
	sensitive("POST /api/register", users.Register)
	sensitive("POST /api/login", users.Login)
	sensitive("POST /api/password/forgot", users.ForgotPassword)
	sensitive("POST /api/password/reset", users.ResetPassword)
	sensitive("POST /api/email/verify", users.VerifyEmail)
	user("POST /api/email/verify/resend", users.ResendVerification)
	user("POST /api/logout-all", users.LogoutEverywhere)
	user("GET /api/profile", users.Profile)
	user("PUT /api/profile", users.UpdateProfile)
	user("PUT /api/profile/password", users.ChangePassword)
	user("DELETE /api/profile", users.Deactivate)
	user("POST /api/profile/image", uploads.ProfileImage)
	user("DELETE /api/profile/image", uploads.DeleteProfileImage)

	// Catalog (public)
	public("GET /api/products", products.List)
	public("GET /api/products/{id}", products.Get)
	public("GET /api/products/{id}/images", products.Images)
	public("GET /api/products/{id}/colors", products.Colors)
	public("GET /api/products/{id}/reviews", reviews.ListForProduct)
	public("GET /api/categories", categories.List)
	public("GET /api/settings", settings.Get)

	// Catalog management (admin)
	admin("GET /api/admin/products", products.AdminList)
	admin("GET /api/admin/products/{id}", products.AdminGet)
	admin("POST /api/products", products.Create)
	admin("PUT /api/products/{id}", products.Update)
	admin("DELETE /api/products/{id}", products.Delete)
	admin("POST /api/products/{id}/images", products.AddImage) // by URL
	admin("POST /api/admin/products/{id}/images/upload", uploads.ProductImage)
	admin("PUT /api/product-images/{id}/primary", products.SetPrimaryImage)
	admin("DELETE /api/product-images/{id}", products.DeleteImage)
	admin("POST /api/products/{id}/colors", products.AddColor)
	admin("DELETE /api/product-colors/{id}", products.DeleteColor)

	// Cart
	user("GET /api/cart", carts.Get)
	user("POST /api/cart/items", carts.AddItem)
	user("PUT /api/cart/items/{id}", carts.UpdateItem)
	user("DELETE /api/cart/items/{id}", carts.RemoveItem)
	user("DELETE /api/cart", carts.Clear)
	userLimited("POST /api/cart/coupon", coupons.Apply)
	user("DELETE /api/cart/coupon", coupons.Remove)

	// Favorites
	user("GET /api/favorites", favorites.List)
	user("POST /api/favorites", favorites.Add)
	user("DELETE /api/favorites/{product_id}", favorites.Remove)

	// Addresses
	user("GET /api/addresses", addresses.List)
	user("POST /api/addresses", addresses.Create)
	user("GET /api/addresses/{id}", addresses.Get)
	user("PUT /api/addresses/{id}", addresses.Update)
	user("DELETE /api/addresses/{id}", addresses.Delete)
	user("PUT /api/addresses/{id}/default", addresses.SetDefault)

	// Orders
	verified("POST /api/orders", orders.Checkout)
	user("GET /api/orders", orders.ListMine)
	user("GET /api/orders/{id}", orders.GetMine)
	user("POST /api/orders/{id}/cancel", orders.CancelMine)
	admin("GET /api/admin/orders", orders.AdminList)
	admin("GET /api/admin/orders/{id}", orders.AdminGet)
	admin("PUT /api/admin/orders/{id}/status", orders.AdminUpdateStatus)
	admin("PUT /api/admin/orders/{id}/note", orders.AdminSetNote)

	// CSV downloads (admin)
	exports := handlers.NewExportHandler(services.NewExportService(
		repository.NewExportRepository(db), cfg.StoreTimezone))
	admin("GET /api/admin/exports/orders.csv", exports.Orders)
	admin("GET /api/admin/exports/customers.csv", exports.Customers)

	// Reviews
	user("POST /api/products/{id}/reviews", reviews.Create)
	user("PUT /api/reviews/{id}", reviews.Update)
	user("DELETE /api/reviews/{id}", reviews.Delete)
	user("POST /api/reviews/{id}/helpful", reviews.AddHelpfulVote)
	user("DELETE /api/reviews/{id}/helpful", reviews.RemoveHelpfulVote)

	// Offers (timed promotions)
	public("GET /api/offers", offers.ListLive)
	public("GET /api/offers/{id}", offers.GetLive)
	admin("GET /api/admin/offers", offers.AdminList)
	admin("GET /api/admin/offers/{id}", offers.AdminGet)
	admin("POST /api/admin/offers", offers.Create)
	admin("PUT /api/admin/offers/{id}", offers.Update)
	admin("DELETE /api/admin/offers/{id}", offers.Delete)

	// Bundles
	public("GET /api/bundles", bundles.List)
	public("GET /api/bundles/{id}", bundles.Get)
	admin("GET /api/admin/bundles", bundles.AdminList)
	admin("GET /api/admin/bundles/{id}", bundles.AdminGet)
	admin("POST /api/admin/bundles", bundles.Create)
	admin("PUT /api/admin/bundles/{id}", bundles.Update)
	admin("DELETE /api/admin/bundles/{id}", bundles.Delete)
	user("POST /api/cart/bundles", carts.AddBundle)
	user("PUT /api/cart/bundles/{id}", carts.UpdateBundle)
	user("DELETE /api/cart/bundles/{id}", carts.RemoveBundle)

	// Customer debts (pay later)
	user("GET /api/debt", debts.Mine)
	admin("GET /api/admin/debts", debts.AdminList)
	admin("GET /api/admin/customers/{id}/debt", debts.AdminStatement)
	admin("PUT /api/admin/customers/{id}/credit-limit", debts.SetCreditLimit)
	admin("POST /api/admin/customers/{id}/debt/payments", debts.RecordPayment)
	admin("POST /api/admin/customers/{id}/debt/adjustments", debts.RecordAdjustment)

	// Waiting list
	user("POST /api/waiting-list", waiting.Join)
	user("GET /api/waiting-list", waiting.ListMine)
	user("DELETE /api/waiting-list/{id}", waiting.CancelMine)
	admin("GET /api/admin/waiting-list", waiting.AdminList)
	admin("PUT /api/admin/waiting-list/{id}/status", waiting.AdminUpdateStatus)

	// Categories, store settings and users (admin)
	admin("GET /api/admin/categories", categories.AdminList)
	admin("POST /api/admin/categories", categories.Create)
	admin("PUT /api/admin/categories/{id}", categories.Update)
	admin("DELETE /api/admin/categories/{id}", categories.Delete)
	admin("PUT /api/admin/settings", settings.Save)
	admin("POST /api/admin/settings/images", uploads.SettingsImage) // banner images (multipart, field image)
	admin("GET /api/admin/users", adminUsers.List)
	admin("POST /api/admin/users", adminUsers.Create)
	admin("GET /api/admin/users/{id}", adminUsers.Get)
	admin("PUT /api/admin/users/{id}", adminUsers.Update)
	admin("DELETE /api/admin/users/{id}", adminUsers.Delete)

	// Custom domains (store admin), and the HTTPS check Caddy asks before issuing a certificate
	domains := handlers.NewDomainHandler(services.NewDomainService(repository.NewDomainRepository(db), resolver, cfg.PlatformDomain,
		func(s tenant.Store) string { return StoreHost(cfg, s) }), resolver, cfg.PlatformAdminHost)
	admin("GET /api/admin/domains", domains.List)
	admin("POST /api/admin/domains", domains.Add)
	admin("POST /api/admin/domains/{id}/verify", domains.Verify)
	admin("POST /api/admin/domains/{id}/primary", domains.SetPrimary)
	admin("DELETE /api/admin/domains/{id}", domains.Delete)
	public("GET /internal/tls-check", domains.TLSCheck)

	// Coupons (admin)
	admin("GET /api/admin/coupons", coupons.AdminList)
	admin("POST /api/admin/coupons", coupons.Create)
	admin("GET /api/admin/coupons/{id}", coupons.AdminGet)
	admin("PUT /api/admin/coupons/{id}", coupons.Update)
	admin("DELETE /api/admin/coupons/{id}", coupons.Delete)

	// Dashboard (admin)
	admin("GET /api/admin/dashboard", dashboard.Get)

	// Notification email queue (admin)
	admin("GET /api/admin/emails", emails.List)
	admin("GET /api/admin/emails/summary", emails.Summary)
	admin("GET /api/admin/emails/{id}", emails.Get)
	admin("POST /api/admin/emails/{id}/retry", emails.Retry)
	admin("POST /api/admin/emails/retry-failed", emails.RetryFailed)
	admin("DELETE /api/admin/emails/{id}", emails.Cancel)

	// ---- platform super admin (only on PLATFORM_ADMIN_HOST, see middleware.Store) ----
	platformTokens := auth.NewPlatformTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	platformSvc := services.NewPlatformService(repository.NewPlatformRepository(db), platformTokens, mailer, tpl,
		func(s tenant.Store) string { return StoreURL(cfg, s) }, resolver, loginFailures)
	platform := handlers.NewPlatformHandler(platformSvc)
	platformAdmin := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, middleware.AuthenticatePlatform(platformTokens, platformSvc)(httpx.Handle(h)))
	}
	sensitive("POST /api/platform/login", platform.Login)
	platformAdmin("GET /api/platform/me", platform.Me)
	platformAdmin("GET /api/platform/stores", platform.ListStores)
	platformAdmin("POST /api/platform/stores", platform.CreateStore)
	platformAdmin("GET /api/platform/stores/{id}", platform.GetStore)
	platformAdmin("PUT /api/platform/stores/{id}", platform.UpdateStore)
	platformAdmin("POST /api/platform/stores/{id}/suspend", platform.Suspend)
	platformAdmin("POST /api/platform/stores/{id}/activate", platform.Activate)
	platformAdmin("POST /api/platform/stores/{id}/invite", platform.ResendInvite)

	// Global middleware, outermost first: recover from panics, log the request,
	// find the client IP, answer CORS (so even a 429 carries CORS headers and
	// the browser can read it), find the store from the host name (everything
	// after this runs for that store only), apply the general per-IP limit,
	// then route.
	var h http.Handler = mux
	h = middleware.RateLimit(general)(h)
	h = middleware.Store(resolver, cfg.TrustProxy, cfg.PlatformAdminHost)(h)
	h = middleware.CORS(cfg.AllowedOrigins)(h)
	h = middleware.RealIP(cfg.TrustProxy)(h)
	h = middleware.SecurityHeaders(cfg.Production())(h) // also on 404/429 and panics
	h = middleware.Logger(h)
	h = middleware.Recover(h)

	worker.Start()
	return &App{Handler: h, Notifications: worker, users: userSvc}, nil
}
