package validator

import (
	"errors"
	"fmt"
	"math"

	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
)

var (
	ErrInvalidSymbol      = errors.New("invalid symbol")
	ErrInvalidSide        = errors.New("side must be BUY or SELL")
	ErrInvalidQuantity    = errors.New("quantity must be positive")
	ErrInvalidPrice       = errors.New("price must be positive")
	ErrSlippageTooHigh    = errors.New("slippage exceeds maximum allowed")
	ErrInvalidStopLoss    = errors.New("invalid stop-loss configuration")
	ErrInvalidTakeProfit  = errors.New("invalid take-profit configuration")
	ErrInsufficientBalance = errors.New("insufficient balance for order")
	ErrExposureLimitExceeded = errors.New("order exceeds maximum exposure ratio")
)

// Config holds validator configuration.
type Config struct {
	DefaultMaxSlippage float64 // e.g., 0.02 = 2%
}

// BalanceChecker defines the interface for checking user balances.
type BalanceChecker interface {
	GetAvailableBalance(userID int64, asset string) (float64, error)
	GetTotalBalance(userID int64) (float64, error)
}

// MockBalanceChecker provides a mock implementation for testing.
type MockBalanceChecker struct {
	Balances      map[int64]map[string]float64
	TotalBalances map[int64]float64
}

// NewMockBalanceChecker creates a new MockBalanceChecker.
func NewMockBalanceChecker() *MockBalanceChecker {
	return &MockBalanceChecker{
		Balances:      make(map[int64]map[string]float64),
		TotalBalances: make(map[int64]float64),
	}
}

func (m *MockBalanceChecker) GetAvailableBalance(userID int64, asset string) (float64, error) {
	if userBalances, ok := m.Balances[userID]; ok {
		if balance, ok := userBalances[asset]; ok {
			return balance, nil
		}
	}
	return 0, nil
}

func (m *MockBalanceChecker) GetTotalBalance(userID int64) (float64, error) {
	if total, ok := m.TotalBalances[userID]; ok {
		return total, nil
	}
	return 0, nil
}

// Validator validates incoming trade signals.
type Validator struct {
	config         Config
	balanceChecker BalanceChecker
}

// New creates a new Validator.
func New(cfg Config, balanceChecker BalanceChecker) *Validator {
	if cfg.DefaultMaxSlippage == 0 {
		cfg.DefaultMaxSlippage = 0.02 // 2% default
	}
	return &Validator{
		config:         cfg,
		balanceChecker: balanceChecker,
	}
}

// ValidationResult holds the outcome of signal validation.
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// ValidateSignal validates a trade signal against all rules.
func (v *Validator) ValidateSignal(signal eventbus.TradeSignal, livePrice float64, maxExposureRatio float64) ValidationResult {
	result := ValidationResult{Valid: true}

	if err := v.validateBasicFields(signal); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
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

	if err := v.validateBalance(signal); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	if err := v.validateExposure(signal, maxExposureRatio); err != nil {
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
	if livePrice <= 0 {
		return nil // skip slippage check if no live price available
	}

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
		return nil // stop-loss is optional
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
		return nil // take-profit is optional
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

func (v *Validator) validateBalance(signal eventbus.TradeSignal) error {
	if v.balanceChecker == nil {
		return nil
	}

	// Determine the quote asset needed (simplified: assume USDT pairs)
	requiredBalance := signal.Quantity * signal.Price

	balance, err := v.balanceChecker.GetAvailableBalance(signal.UserID, "USDT")
	if err != nil {
		return fmt.Errorf("failed to check balance: %w", err)
	}

	if balance < requiredBalance {
		return fmt.Errorf("%w: need %.2f USDT, have %.2f", ErrInsufficientBalance, requiredBalance, balance)
	}
	return nil
}

func (v *Validator) validateExposure(signal eventbus.TradeSignal, maxExposureRatio float64) error {
	if v.balanceChecker == nil || maxExposureRatio == 0 {
		return nil
	}

	totalBalance, err := v.balanceChecker.GetTotalBalance(signal.UserID)
	if err != nil {
		return fmt.Errorf("failed to check total balance: %w", err)
	}

	if totalBalance == 0 {
		return nil // skip if no balance info available
	}

	orderValue := signal.Quantity * signal.Price
	exposureRatio := orderValue / totalBalance

	if exposureRatio > maxExposureRatio {
		return fmt.Errorf("%w: order %.4f exceeds max %.4f", ErrExposureLimitExceeded, exposureRatio, maxExposureRatio)
	}
	return nil
}
