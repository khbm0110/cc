package validator

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/khbm0110/copy-trading-platform/internal/binance"
	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
)

var (
	ErrInvalidSymbol         = errors.New("invalid symbol")
	ErrInvalidSide           = errors.New("side must be BUY or SELL")
	ErrInvalidQuantity       = errors.New("quantity must be positive")
	ErrInvalidPrice          = errors.New("price must be positive")
	ErrSlippageTooHigh       = errors.New("slippage exceeds maximum allowed")
	ErrInvalidStopLoss       = errors.New("invalid stop-loss configuration")
	ErrInvalidTakeProfit     = errors.New("invalid take-profit configuration")
	ErrInsufficientBalance   = errors.New("insufficient balance for order")
	ErrExposureLimitExceeded = errors.New("order exceeds maximum exposure ratio")
)

type Config struct {
	DefaultMaxSlippage float64
}

type BalanceChecker interface {
	GetAvailableBalance(ctx context.Context, client binance.Client, asset string) (float64, error)
}

type LiveBalanceChecker struct{}

func (l *LiveBalanceChecker) GetAvailableBalance(ctx context.Context, client binance.Client, asset string) (float64, error) {
	return client.GetBalance(ctx, asset)
}

// MockBalanceChecker implements BalanceChecker for testing.
// Supports per-user balances via map[userID]map[asset]balance.
// Returns explicit error when asset is not found.
type MockBalanceChecker struct {
	mu       sync.RWMutex
	balances map[int64]map[string]float64
}

// NewMockBalanceChecker creates a new MockBalanceChecker.
func NewMockBalanceChecker() *MockBalanceChecker {
	return &MockBalanceChecker{
		balances: make(map[int64]map[string]float64),
	}
}

// SetBalance sets the balance for a specific user and asset.
// This method is used for test setup.
func (m *MockBalanceChecker) SetBalance(userID int64, asset string, amount float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances[userID] == nil {
		m.balances[userID] = make(map[string]float64)
	}
	m.balances[userID][asset] = amount
}

// SetUserBalances sets all balances for a user at once.
func (m *MockBalanceChecker) SetUserBalances(userID int64, balances map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balances[userID] = balances
}

// GetAvailableBalance returns the balance for the specified user and asset.
// Returns explicit error if asset not found for the user.
func (m *MockBalanceChecker) GetAvailableBalance(ctx context.Context, client binance.Client, asset string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For MockBalanceChecker, we use client.UserID if available
	// In real tests, you'd pass the userID through context or as part of the mock setup
	userID := int64(1) // Default user ID for testing

	userBalances, ok := m.balances[userID]
	if !ok {
		return 0, fmt.Errorf("asset %s not found for user %d", asset, userID)
	}

	balance, ok := userBalances[asset]
	if !ok {
		return 0, fmt.Errorf("asset %s not found for user %d", asset, userID)
	}

	return balance, nil
}

// GetAvailableBalanceForUser returns the balance for a specific user and asset.
// This method allows setting userID explicitly.
func (m *MockBalanceChecker) GetAvailableBalanceForUser(ctx context.Context, userID int64, asset string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userBalances, ok := m.balances[userID]
	if !ok {
		return 0, fmt.Errorf("asset %s not found for user %d", asset, userID)
	}

	balance, ok := userBalances[asset]
	if !ok {
		return 0, fmt.Errorf("asset %s not found for user %d", asset, userID)
	}

	return balance, nil
}

type Validator struct {
	config         Config
	balanceChecker BalanceChecker
}

func New(cfg Config, balanceChecker BalanceChecker) *Validator {
	if cfg.DefaultMaxSlippage == 0 {
		cfg.DefaultMaxSlippage = 0.02
	}
	return &Validator{
		config:         cfg,
		balanceChecker: balanceChecker,
	}
}

type ValidationResult struct {
	Valid  bool
	Errors []string
}

func (v *Validator) ValidateSignal(ctx context.Context, client binance.Client, signal eventbus.TradeSignal, maxExposureRatio float64) ValidationResult {
	result := ValidationResult{Valid: true}

	if err := v.validateBasicFields(signal); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	livePrice, err := client.GetTickerPrice(ctx, signal.Symbol)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to fetch live price: %v", err))
		return result
	}

	if err := v.validateSlippage(signal, livePrice); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	if err := v.validateStopLoss(signal); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	if err := v.validateTakeProfit(signal); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	if err := v.validateBalance(ctx, client, signal); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	if err := v.validateExposureRatio(ctx, client, signal, maxExposureRatio); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	return result
}

func (v *Validator) validateBasicFields(signal eventbus.TradeSignal) error {
	if signal.Symbol == "" {
		return ErrInvalidSymbol
	}
	if signal.Side != "BUY" && signal.Side != "SELL" {
		return ErrInvalidSide
	}
	if signal.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if signal.Price <= 0 {
		return ErrInvalidPrice
	}
	if signal.ClientOrderID == "" {
		return fmt.Errorf("client_order_id is required")
	}
	return nil
}

func (v *Validator) validateSlippage(signal eventbus.TradeSignal, livePrice float64) error {
	maxSlippage := signal.MaxSlippage
	if maxSlippage == 0 {
		maxSlippage = v.config.DefaultMaxSlippage
	}

	slippage := math.Abs(signal.Price-livePrice) / livePrice
	if slippage > maxSlippage {
		return fmt.Errorf("%w: %.4f%% (max: %.4f%%)", ErrSlippageTooHigh, slippage*100, maxSlippage*100)
	}
	return nil
}

func (v *Validator) validateStopLoss(signal eventbus.TradeSignal) error {
	if signal.StopLoss == 0 {
		return nil
	}
	if signal.Side == "BUY" {
		if signal.StopLoss >= signal.Price {
			return fmt.Errorf("%w: stop-loss must be below entry price for BUY", ErrInvalidStopLoss)
		}
	} else {
		if signal.StopLoss <= signal.Price {
			return fmt.Errorf("%w: stop-loss must be above entry price for SELL", ErrInvalidStopLoss)
		}
	}
	return nil
}

func (v *Validator) validateTakeProfit(signal eventbus.TradeSignal) error {
	if signal.TakeProfit == 0 {
		return nil
	}
	if signal.Side == "BUY" {
		if signal.TakeProfit <= signal.Price {
			return fmt.Errorf("%w: take-profit must be above entry price for BUY", ErrInvalidTakeProfit)
		}
	} else {
		if signal.TakeProfit >= signal.Price {
			return fmt.Errorf("%w: take-profit must be below entry price for SELL", ErrInvalidTakeProfit)
		}
	}
	return nil
}

func (v *Validator) validateBalance(ctx context.Context, client binance.Client, signal eventbus.TradeSignal) error {
	if v.balanceChecker == nil {
		return nil // Skip balance check if no checker provided
	}

	requiredBalance := signal.Quantity * signal.Price
	balance, err := v.balanceChecker.GetAvailableBalance(ctx, client, "USDT")
	if err != nil {
		return fmt.Errorf("failed to check balance: %w", err)
	}
	if balance < requiredBalance {
		return fmt.Errorf("%w: need %.2f USDT, have %.2f", ErrInsufficientBalance, requiredBalance, balance)
	}
	return nil
}

func (v *Validator) validateExposureRatio(ctx context.Context, client binance.Client, signal eventbus.TradeSignal, maxExposureRatio float64) error {
	if v.balanceChecker == nil || maxExposureRatio <= 0 {
		return nil // Skip exposure check if no balance checker or no limit
	}

	orderValue := signal.Quantity * signal.Price
	balance, err := v.balanceChecker.GetAvailableBalance(ctx, client, "USDT")
	if err != nil {
		return fmt.Errorf("failed to check balance for exposure: %w", err)
	}

	if balance <= 0 {
		return fmt.Errorf("%w: cannot calculate exposure with zero balance", ErrExposureLimitExceeded)
	}

	exposureRatio := orderValue / balance
	if exposureRatio > maxExposureRatio {
		return fmt.Errorf("%w: %.2f%% (max: %.2f%%)", ErrExposureLimitExceeded, exposureRatio*100, maxExposureRatio*100)
	}

	return nil
}
