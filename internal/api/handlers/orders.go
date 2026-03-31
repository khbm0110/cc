package handlers

import (
	"log/slog"
	"net/http"

	"github.com/khbm0110/copy-trading-platform/internal/order"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

type OrdersHandler struct {
	orderRepo order.Repository
	userRepo  user.Repository
	logger    *slog.Logger
}

func NewOrdersHandler(orderRepo order.Repository, userRepo user.Repository, logger *slog.Logger) *OrdersHandler {
	return &OrdersHandler{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		logger:    logger.With(slog.String("component", "orders_handler")),
	}
}

func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte("not implemented"))
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte("not implemented"))
}

func (h *OrdersHandler) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte("not implemented"))
}