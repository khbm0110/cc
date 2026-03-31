package validator

import (
	"context"
	"testing"

	"github.com/khbm0110/copy-trading-platform/internal/binance"
	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
)

// mockBinanceClient implements binance.Client for testing
type mockBinanceClient struct {
	prices   map[string]float64
	balances map[string]float64
}

func (m *mockBinanceClient) GetTickerPrice(ctx context.Context, symbol string) (float64, error) {
	if price, ok := m.prices[symbol]; ok {
		return price, nil
	}
	return 0, nil
}

func (m *mockBinanceClient) GetBalance(ctx context.Context, asset string) (float64, error) {
	if balance, ok := m.balances[asset]; ok {
		return balance, nil
	}
	return 0, nil
}

func (m *mockBinanceClient) ExecuteOrder(ctx context.Context, req binance.OrderRequest) (*binance.OrderResponse, error) {
	return &binance.OrderResponse{
		Symbol:        req.Symbol,
		ClientOrderID: req.ClientOrderID,
		Status:        "NEW",
	}, nil
}

func (m *mockBinanceClient) QueryOrderStatus(ctx context.Context, symbol string, clientOrderID string) (*binance.QueryOrderResponse, error) {
	return &binance.QueryOrderResponse{
		Symbol:        symbol,
		ClientOrderID: clientOrderID,
		Status:        "FILLED",
	}, nil
}

func TestValidateSignal_ValidBuySignal(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.SetUserBalances(1, map[string]float64{"USDT": 10000})

	v := New(Config{DefaultMaxSlippage: 0.02}, bc)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-1",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      0.1,
		Price:         50000,
		StopLoss:      49000,
		TakeProfit:    52000,
		ClientOrderID: "ord-1",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.5)
	if !result.Valid {
		t.Errorf("expected valid signal, got errors: %v", result.Errors)
	}
}

func TestValidateSignal_ValidSellSignal(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.SetUserBalances(1, map[string]float64{"USDT": 10000})

	v := New(Config{DefaultMaxSlippage: 0.02}, bc)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-2",
		UserID:        1,
		Symbol:        "ETHUSDT",
		Side:          "SELL",
		Quantity:      1.0,
		Price:         3000,
		StopLoss:      3100,
		TakeProfit:    2800,
		ClientOrderID: "ord-2",
	}

	client := &mockBinanceClient{prices: map[string]float64{"ETHUSDT": 3000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.5)
	if !result.Valid {
		t.Errorf("expected valid signal, got errors: %v", result.Errors)
	}
}

func TestValidateSignal_InvalidSymbol(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-3",
		UserID:        1,
		Symbol:        "",
		Side:          "BUY",
		Quantity:      1,
		Price:         100,
		ClientOrderID: "ord-3",
	}

	client := &mockBinanceClient{prices: map[string]float64{}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for empty symbol")
	}
}

func TestValidateSignal_InvalidSide(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-4",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "HOLD",
		Quantity:      1,
		Price:         100,
		ClientOrderID: "ord-4",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 100}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for bad side")
	}
}

func TestValidateSignal_NegativeQuantity(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-5",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      -1,
		Price:         100,
		ClientOrderID: "ord-5",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 100}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for negative quantity")
	}
}

func TestValidateSignal_SlippageTooHigh(t *testing.T) {
	v := New(Config{DefaultMaxSlippage: 0.01}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-6",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000,
		ClientOrderID: "ord-6",
	}

	// Live price 47500 (5% slippage), max 1%
	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 47500}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for high slippage")
	}
}

func TestValidateSignal_SlippageWithinLimit(t *testing.T) {
	v := New(Config{DefaultMaxSlippage: 0.05}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-7",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000,
		ClientOrderID: "ord-7",
	}

	// 1% slippage, max 5%
	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 49500}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid == false {
		t.Errorf("expected valid signal, slippage within limit, got errors: %v", result.Errors)
	}
}

func TestValidateSignal_InvalidStopLossBuy(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-8",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000,
		StopLoss:      51000, // above entry for BUY = invalid
		ClientOrderID: "ord-8",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for bad stop-loss on BUY")
	}
}

func TestValidateSignal_InvalidStopLossSell(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-9",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "SELL",
		Quantity:      1,
		Price:         50000,
		StopLoss:      49000, // below entry for SELL = invalid
		ClientOrderID: "ord-9",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for bad stop-loss on SELL")
	}
}

func TestValidateSignal_InvalidTakeProfitBuy(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-10",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000,
		TakeProfit:    49000, // below entry for BUY = invalid
		ClientOrderID: "ord-10",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for bad take-profit on BUY")
	}
}

func TestValidateSignal_InvalidTakeProfitSell(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-11",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "SELL",
		Quantity:      1,
		Price:         50000,
		TakeProfit:    51000, // above entry for SELL = invalid
		ClientOrderID: "ord-11",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for bad take-profit on SELL")
	}
}

func TestValidateSignal_InsufficientBalance(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.SetUserBalances(1, map[string]float64{"USDT": 100})

	v := New(Config{}, bc)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-12",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000, // needs 50000 USDT, has 100
		ClientOrderID: "ord-12",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 1.0)
	if result.Valid {
		t.Error("expected invalid signal for insufficient balance")
	}
}

func TestValidateSignal_ExposureLimitExceeded(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.SetUserBalances(1, map[string]float64{"USDT": 100000})

	v := New(Config{}, bc)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-13",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000, // 50% exposure, max 10%
		ClientOrderID: "ord-13",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.10)
	if result.Valid {
		t.Error("expected invalid signal for exposure limit exceeded")
	}
}

func TestValidateSignal_MissingClientOrderID(t *testing.T) {
	v := New(Config{}, nil)

	signal := eventbus.TradeSignal{
		SignalID: "sig-14",
		UserID:   1,
		Symbol:   "BTCUSDT",
		Side:     "BUY",
		Quantity: 1,
		Price:    50000,
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for missing client_order_id")
	}
}

func TestValidateSignal_NoBalanceChecker(t *testing.T) {
	v := New(Config{DefaultMaxSlippage: 0.05}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-15",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000,
		ClientOrderID: "ord-15",
	}

	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 50000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if !result.Valid {
		t.Errorf("expected valid signal without balance checker, got errors: %v", result.Errors)
	}
}

func TestValidateSignal_CustomMaxSlippage(t *testing.T) {
	v := New(Config{DefaultMaxSlippage: 0.01}, nil)

	signal := eventbus.TradeSignal{
		SignalID:      "sig-16",
		UserID:        1,
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000,
		MaxSlippage:   0.05, // custom 5% max slippage
		ClientOrderID: "ord-16",
	}

	// 2% slippage, which exceeds default 1% but within custom 5%
	client := &mockBinanceClient{prices: map[string]float64{"BTCUSDT": 49000}}
	result := v.ValidateSignal(context.Background(), client, signal, 0.1)
	if !result.Valid {
		t.Errorf("expected valid signal with custom max slippage, got errors: %v", result.Errors)
	}
}

// Test MockBalanceChecker methods
func TestMockBalanceChecker_SetBalance(t *testing.T) {
	bc := NewMockBalanceChecker()

	bc.SetBalance(1, "USDT", 1000.50)
	bc.SetBalance(1, "BTC", 0.5)
	bc.SetBalance(2, "USDT", 500.0)

	balance, err := bc.GetAvailableBalanceForUser(context.Background(), 1, "USDT")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if balance != 1000.50 {
		t.Errorf("expected 1000.50, got %f", balance)
	}

	balance, err = bc.GetAvailableBalanceForUser(context.Background(), 1, "BTC")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if balance != 0.5 {
		t.Errorf("expected 0.5, got %f", balance)
	}

	balance, err = bc.GetAvailableBalanceForUser(context.Background(), 2, "USDT")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if balance != 500.0 {
		t.Errorf("expected 500.0, got %f", balance)
	}
}

func TestMockBalanceChecker_AssetNotFound(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.SetBalance(1, "USDT", 1000)

	_, err := bc.GetAvailableBalanceForUser(context.Background(), 1, "ETH")
	if err == nil {
		t.Error("expected error for missing asset")
	}
}

func TestMockBalanceChecker_UserNotFound(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.SetBalance(1, "USDT", 1000)

	_, err := bc.GetAvailableBalanceForUser(context.Background(), 999, "USDT")
	if err == nil {
		t.Error("expected error for missing user")
	}
}
