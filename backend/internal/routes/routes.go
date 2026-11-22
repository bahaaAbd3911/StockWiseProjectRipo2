// Package routes builds the whole application: it creates every repository,
// service and handler (dependency injection), maps URLs to handlers, and wraps
// everything in the global middleware. Reading this file top to bottom shows
// every endpoint the API offers and who may call it.
package routes

import (
	"database/sql"
	"net/http"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/handlers"
	"ecommerce-backend/internal/httpx"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/services"
)

// New returns the root HTTP handler.
func New(db *sql.DB, cfg config.Config) http.Handler {
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
	userSvc := services.NewUserService(userRepo, tokens)
	productSvc := services.NewProductService(productRepo, catalogRepo)
	cartSvc := services.NewCartService(cartRepo)
	favoriteSvc := services.NewFavoriteService(favoriteRepo)
	addressSvc := services.NewAddressService(addressRepo)
	orderSvc := services.NewOrderService(orderRepo)
	reviewSvc := services.NewReviewService(reviewRepo, catalogRepo)
	waitingSvc := services.NewWaitingListService(waitingRepo)
	offerSvc := services.NewOfferService(offerRepo, productRepo)
	bundleSvc := services.NewBundleService(bundleRepo)
	debtSvc := services.NewDebtService(debtRepo)

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

	// ---- access levels ----
	mux := http.NewServeMux()
	authenticate := middleware.Authenticate(tokens, userRepo)

	public := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, httpx.Handle(h))
	}
	user := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, authenticate(httpx.Handle(h)))
	}
	admin := func(pattern string, h httpx.HandlerFunc) {
		mux.Handle(pattern, authenticate(middleware.RequireAdmin(httpx.Handle(h))))
	}

	// Health
	public("GET /{$}", health.Root)
	public("GET /api/health", health.Health)

	// Account
	public("POST /api/register", users.Register)
	public("POST /api/login", users.Login)
	user("GET /api/profile", users.Profile)
	user("PUT /api/profile", users.UpdateProfile)
	user("PUT /api/profile/password", users.ChangePassword)
	user("DELETE /api/profile", users.Deactivate)

	// Catalog (public)
	public("GET /api/products", products.List)
	public("GET /api/products/{id}", products.Get)
	public("GET /api/products/{id}/images", products.Images)
	public("GET /api/products/{id}/colors", products.Colors)
	public("GET /api/products/{id}/reviews", reviews.ListForProduct)
	public("GET /api/categories", products.Categories)

	// Catalog management (admin)
	admin("GET /api/admin/products", products.AdminList)
	admin("GET /api/admin/products/{id}", products.AdminGet)
	admin("POST /api/products", products.Create)
	admin("PUT /api/products/{id}", products.Update)
	admin("DELETE /api/products/{id}", products.Delete)
	admin("POST /api/products/{id}/images", products.AddImage)
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
	user("POST /api/orders", orders.Checkout)
	user("GET /api/orders", orders.ListMine)
	user("GET /api/orders/{id}", orders.GetMine)
	user("POST /api/orders/{id}/cancel", orders.CancelMine)
	admin("GET /api/admin/orders", orders.AdminList)
	admin("GET /api/admin/orders/{id}", orders.AdminGet)
	admin("PUT /api/admin/orders/{id}/status", orders.AdminUpdateStatus)

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

	// Global middleware, outermost first: recover from panics, log the request,
	// then answer CORS before routing.
	return middleware.Recover(middleware.Logger(middleware.CORS(cfg.AllowedOrigins)(mux)))
}
