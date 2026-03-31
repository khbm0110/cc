package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/khbm0110/copy-trading-platform/internal/api/middleware"
	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

type SignalsHandler struct {
	eventBus *eventbus.EventBus
	userRepo user.Repository
	logger   *slog.Logger
}

func NewSignalsHandler(eventBus *eventbus.EventBus, userRepo user.Repository, logger *slog.Logger) *SignalsHandler {
	return &SignalsHandler{
		eventBus: eventBus,
		userRepo: userRepo,
		logger:   logger.With(slog.String("component", "signals_handler")),
	}
}

func (h *SignalsHandler) CreateSignal(w http.ResponseWriter, r *http.Request) {
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

	if u.Role != "trader" {
		http.Error(w, "forbidden: trader role required", http.StatusForbidden)
		return
	}

	var signal eventbus.TradeSignal
	if err := json.NewDecoder(r.Body).Decode(&signal); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Generate SignalID as UUID
	signalID, err := generateUUID()
	if err != nil {
		h.logger.Error("failed to generate signal id", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	signal.SignalID = signalID

	// Generate ClientOrderID as UUID
	clientOrderID, err := generateUUID()
	if err != nil {
		h.logger.Error("failed to generate client order id", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	signal.ClientOrderID = clientOrderID

	// Set UserID from context
	signal.UserID = userID

	// Publish to eventbus
	_, err = h.eventBus.Publish(ctx, "trade_signals", signal)
	if err != nil {
		h.logger.Error("failed to publish signal", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("signal created",
		"signal_id", signal.SignalID,
		"user_id", userID,
		"symbol", signal.Symbol,
		"side", signal.Side,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(signal)
}

func generateUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
