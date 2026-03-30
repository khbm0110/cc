package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/khbm0110/copy-trading-platform/internal/auth"
	"github.com/khbm0110/copy-trading-platform/internal/user"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	User         user.User `json:"user"`
}

type AuthHandler struct {
	repo   user.Repository
	secret string
	logger *slog.Logger
}

func NewAuthHandler(repo user.Repository, secret string, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		repo:   repo,
		secret: secret,
		logger: logger.With(slog.String("component", "auth_handler")),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "email, password, and name are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("failed to hash password", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	u, err := h.repo.CreateUser(r.Context(), user.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		Role:         "client",
	})
	if err != nil {
		h.logger.Error("failed to create user", slog.String("error", err.Error()))
		http.Error(w, "could not create user", http.StatusConflict)
		return
	}

	token, err := auth.GenerateToken(u.ID, u.Email, u.Role, h.secret, 24)
	if err != nil {
		h.logger.Error("failed to generate token", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: *u})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	u, err := h.repo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, u.PasswordHash) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(u.ID, u.Email, u.Role, h.secret, 24)
	if err != nil {
		h.logger.Error("failed to generate token", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		h.logger.Error("failed to generate refresh token", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.repo.UpdateRefreshToken(r.Context(), u.ID, refreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
		h.logger.Error("failed to store refresh token", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *u,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	u, err := h.repo.GetByRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(u.ID, u.Email, u.Role, h.secret, 24)
	if err != nil {
		h.logger.Error("failed to generate token", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: *u})
}