package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/khbm0110/copy-trading-platform/internal/api/middleware"
	"github.com/khbm0110/copy-trading-platform/internal/binance"
	"github.com/khbm0110/copy-trading-platform/internal/config"
	"github.com/khbm0110/copy-trading-platform/internal/kms"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

type UsersHandler struct {
	userRepo user.Repository
	kms      kms.KeyManager
	cfg      *config.Config
	logger   *slog.Logger
}

func NewUsersHandler(repo user.Repository, kms kms.KeyManager, cfg *config.Config, logger *slog.Logger) *UsersHandler {
	return &UsersHandler{
		userRepo: repo,
		kms:      kms,
		cfg:      cfg,
		logger:   logger,
	}
}

type connectBinanceRequest struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

type userResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type balanceResponse struct {
	Asset string  `json:"asset"`
	Free  float64 `json:"free"`
	Lock  float64 `json:"locked"`
}

func (h *UsersHandler) ConnectBinance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req connectBinanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate keys with Binance
	client := binance.NewRealClient(binance.ClientConfig{
		APIKey:    string(req.APIKey),
		SecretKey: string(req.SecretKey),
		BaseURL:   h.cfg.BinanceAPIURL,
		UserID:    userID,
	}, h.logger)
	_, err := client.GetBalance(ctx, "USDT")
	if err != nil {
		h.logger.Warn("invalid binance keys provided", "user_id", userID, "error", err)
		http.Error(w, "invalid Binance API keys", http.StatusBadRequest)
		return
	}

	// Encrypt keys
	encAPI, err := h.kms.Encrypt(ctx, []byte(req.APIKey))
	if err != nil {
		h.logger.Error("failed to encrypt api key", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	encSecret, err := h.kms.Encrypt(ctx, []byte(req.SecretKey))
	if err != nil {
		h.logger.Error("failed to encrypt secret key", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Update user record
	err = h.userRepo.UpdateAPIKeys(ctx, userID, string(encAPI), string(encSecret))
	if err != nil {
		h.logger.Error("failed to update user api keys", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.logger.Error("failed to fetch user", "user_id", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UsersHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.logger.Error("failed to fetch user", "user_id", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(u.APIKeyEncrypted) == 0 || len(u.SecretKeyEncrypted) == 0 {
		http.Error(w, "binance account not connected", http.StatusBadRequest)
		return
	}

	apiKey, err := h.kms.Decrypt(ctx, u.APIKeyEncrypted)
	if err != nil {
		h.logger.Error("failed to decrypt api key", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	secretKey, err := h.kms.Decrypt(ctx, u.SecretKeyEncrypted)
	if err != nil {
		h.logger.Error("failed to decrypt secret key", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	client := binance.NewRealClient(binance.ClientConfig{
		APIKey:    string(apiKey),
		SecretKey: string(secretKey),
		BaseURL:   h.cfg.BinanceAPIURL,
		UserID:    userID,
	}, h.logger)
	balance, err := client.GetBalance(ctx, "USDT")
	if err != nil {
		h.logger.Error("failed to fetch balance from binance", "error", err)
		http.Error(w, "failed to fetch balance", http.StatusBadGateway)
		return
	}

	resp := balanceResponse{
		Asset: "USDT",
		Free:  balance,
		Lock:  0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}