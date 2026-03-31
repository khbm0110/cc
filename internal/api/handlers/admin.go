package handlers

import (
	"log/slog"
	"net/http"

	"github.com/khbm0110/copy-trading-platform/internal/user"
)

type AdminHandler struct {
	userRepo user.Repository
	logger   *slog.Logger
}

func NewAdminHandler(repo user.Repository, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		userRepo: repo,
		logger:   logger.With(slog.String("component", "admin_handler")),
	}
}

// GetUsers handles GET /api/v1/admin/users
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("501 - Not Implemented"))
}

// ApproveUser handles POST /api/v1/admin/users/{id}/approve
func (h *AdminHandler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("501 - Not Implemented"))
}