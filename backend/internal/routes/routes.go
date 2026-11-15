package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"ecommerce-backend/internal/handlers"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/services"
)

func RegisterRoutes(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /{$}", handlers.HealthCheck)

	// Products
	productRepository := repository.NewProductRepository(db)
	productService := services.NewProductService(productRepository)
	productHandler := handlers.NewProductHandler(productService)

	// Users
	userRepository := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
	userHandler := handlers.NewUserHandler(userService)

	//Favorites
	favoriteRepository := repository.NewFavoriteRepository(db)
	favoriteService := services.NewFavoriteService(favoriteRepository)
	favoriteHandler := handlers.NewFavoriteHandler(favoriteService)

	//Addresses
	addressRepository := repository.NewAddressRepository(db)
	addressService := services.NewAddressService(addressRepository)
	addressHandler := handlers.NewAddressHandler(addressService)

	//cart
	cartRepository := repository.NewCartRepository(db)
	cartService := services.NewCartService(cartRepository)
	cartHandler := handlers.NewCartHandler(cartService)

	//Orders
	orderRepository := repository.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository)
	orderHandler := handlers.NewOrderHandler(orderService)

	// User routes
	mux.HandleFunc("POST /api/register", userHandler.RegisterUser)
	mux.HandleFunc("POST /api/login", userHandler.LoginUser)

	mux.Handle("GET /api/profile", middleware.AuthMiddleware(db,
		http.HandlerFunc(userHandler.GetUserProfile),
	))

	mux.Handle("PUT /api/profile", middleware.AuthMiddleware(db,
		http.HandlerFunc(userHandler.GetUserProfile),
	))

	mux.Handle("PUT /api/profile/password", middleware.AuthMiddleware(db,
		http.HandlerFunc(userHandler.ChangePassword),
	))

	mux.Handle("DELETE /api/profile", middleware.AuthMiddleware(db,
		http.HandlerFunc(userHandler.DeactivateUser),
	))

	mux.Handle("GET /api/protected", middleware.AuthMiddleware(db,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(middleware.UserIDKey).(int64)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "You are authenticated!",
				"user_id": userID,
			})
		}),
	))

	// Product routes
	mux.HandleFunc("GET /api/products", productHandler.GetAllProducts)
	mux.HandleFunc("GET /api/products/{id}", productHandler.GetProductByID)
	mux.HandleFunc("POST /api/products", productHandler.CreateProduct)
	mux.HandleFunc("PUT /api/products/{id}", productHandler.UpdateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", productHandler.DeleteProduct)

	// Product Reviews
	productReviewRepository := repository.NewProductReviewRepository(db)
	productReviewService := services.NewProductReviewService(productReviewRepository)
	productReviewHandler := handlers.NewProductReviewHandler(productReviewService)

	mux.Handle("POST /api/cart/items", middleware.AuthMiddleware(db,
		http.HandlerFunc(cartHandler.AddItem),
	))

	mux.Handle("GET /api/cart", middleware.AuthMiddleware(db,
		http.HandlerFunc(cartHandler.GetCart),
	))

	mux.Handle("PUT /api/cart/items/{id}", middleware.AuthMiddleware(db,
		http.HandlerFunc(cartHandler.UpdateItemQuantity),
	))

	mux.Handle("DELETE /api/cart/items/{id}", middleware.AuthMiddleware(db,
		http.HandlerFunc(cartHandler.RemoveItem),
	))

	mux.Handle("DELETE /api/cart", middleware.AuthMiddleware(db,
		http.HandlerFunc(cartHandler.ClearCart),
	))

	mux.Handle("GET /api/favorites", middleware.AuthMiddleware(db,
		http.HandlerFunc(favoriteHandler.GetFavorites),
	))

	mux.Handle("POST /api/favorites", middleware.AuthMiddleware(db,
		http.HandlerFunc(favoriteHandler.AddFavorite),
	))

	mux.Handle("DELETE /api/favorites/{product_id}", middleware.AuthMiddleware(db,
		http.HandlerFunc(favoriteHandler.RemoveFavorite),
	))

	mux.Handle("GET /api/addresses", middleware.AuthMiddleware(db,
		http.HandlerFunc(addressHandler.GetAddresses),
	))

	mux.Handle("POST /api/addresses", middleware.AuthMiddleware(db,
		http.HandlerFunc(addressHandler.CreateAddress),
	))

	mux.Handle("PUT /api/addresses/{id}", middleware.AuthMiddleware(db,
		http.HandlerFunc(addressHandler.UpdateAddress),
	))

	mux.Handle("DELETE /api/addresses/{id}", middleware.AuthMiddleware(db,
		http.HandlerFunc(addressHandler.DeleteAddress),
	))

	mux.Handle("PUT /api/addresses/{id}/default", middleware.AuthMiddleware(db,
		http.HandlerFunc(addressHandler.SetDefaultAddress),
	))

	mux.Handle("POST /api/orders", middleware.AuthMiddleware(db,
		http.HandlerFunc(orderHandler.CreateOrder),
	))

	mux.Handle("GET /api/orders", middleware.AuthMiddleware(db,
		http.HandlerFunc(orderHandler.GetOrders),
	))

	mux.Handle("GET /api/orders/{id}", middleware.AuthMiddleware(db,
		http.HandlerFunc(orderHandler.GetOrderByID),
	))

	mux.Handle("POST /api/orders/{id}/cancel", middleware.AuthMiddleware(db,
		http.HandlerFunc(orderHandler.CancelOrder),
	))

	mux.Handle("PUT /api/admin/orders/{id}/status",
		middleware.AuthMiddleware(db,
			middleware.AdminMiddleware(db,
				http.HandlerFunc(orderHandler.UpdateOrderStatus),
			),
		),
	)

	// Product Review routes

	// Anyone can view reviews
	mux.HandleFunc("GET /api/products/{id}/reviews",
		productReviewHandler.GetReviewsByProductID,
	)

	// Only authenticated users can create reviews
	mux.Handle("POST /api/products/{id}/reviews",
		middleware.AuthMiddleware(db,
			http.HandlerFunc(productReviewHandler.CreateReview),
		),
	)

	// Only authenticated users can update their own reviews
	mux.Handle("PUT /api/reviews/{id}",
		middleware.AuthMiddleware(db,
			http.HandlerFunc(productReviewHandler.UpdateReview),
		),
	)

	// Only authenticated users can delete their own reviews
	mux.Handle("DELETE /api/reviews/{id}",
		middleware.AuthMiddleware(db,
			http.HandlerFunc(productReviewHandler.DeleteReview),
		),
	)
	
	// Helpful vote routes

	mux.Handle("POST /api/reviews/{id}/helpful",
		middleware.AuthMiddleware(db,
			http.HandlerFunc(productReviewHandler.AddHelpfulVote),
		),
	)

	mux.Handle("DELETE /api/reviews/{id}/helpful",
		middleware.AuthMiddleware(db,
			http.HandlerFunc(productReviewHandler.RemoveHelpfulVote),
		),
	)
	return mux
}
