package api

import (
	"net/http"

	"github.com/khbm0110/copy-trading-platform/internal/api/handlers"
	"github.com/khbm0110/copy-trading-platform/internal/api/middleware"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

func NewRouter(
	authHandler *handlers.AuthHandler,
	usersHandler *handlers.UsersHandler,
	ordersHandler *handlers.OrdersHandler,
	adminHandler *handlers.AdminHandler,
	userRepo user.Repository,
	jwtSecret string,
) http.Handler {
	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Protected Routes setup
	jwtMiddleware := middleware.JWTMiddleware(jwtSecret)

	// Admin Middleware initialization
	adminMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := middleware.GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			u, err := userRepo.GetByID(r.Context(), userID)
			if err != nil || u.Role != "admin" {
				http.Error(w, "forbidden: admin access required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Protected Handlers
	// Auth
	mux.Handle("POST /api/v1/auth/refresh", jwtMiddleware(http.HandlerFunc(authHandler.Refresh)))

	// Users
	mux.Handle("POST /api/v1/users/connect-binance", jwtMiddleware(http.HandlerFunc(usersHandler.ConnectBinance)))
	mux.Handle("GET /api/v1/users/me", jwtMiddleware(http.HandlerFunc(usersHandler.GetMe)))
	mux.Handle("GET /api/v1/users/me/balance", jwtMiddleware(http.HandlerFunc(usersHandler.GetBalance)))

	// Orders
	mux.Handle("POST /api/v1/orders", jwtMiddleware(http.HandlerFunc(ordersHandler.CreateOrder)))
	mux.Handle("GET /api/v1/orders", jwtMiddleware(http.HandlerFunc(ordersHandler.GetOrders)))
	mux.Handle("GET /api/v1/orders/{id}", jwtMiddleware(http.HandlerFunc(ordersHandler.GetOrderDetails)))

	// Admin
	mux.Handle("GET /api/v1/admin/users", jwtMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.GetUsers))))
	mux.Handle("POST /api/v1/admin/users/{id}/approve", jwtMiddleware(adminMiddleware(http.HandlerFunc(adminHandler.ApproveUser))))

	// Global 404
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return mux
}