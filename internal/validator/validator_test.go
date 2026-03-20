package validator

import (
	"testing"

	"github.com/khbm0110/copy-trading-platform/internal/eventbus"
)

func TestValidateSignal_ValidBuySignal(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.Balances[1] = map[string]float64{"USDT": 10000}
	bc.TotalBalances[1] = 10000

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

	result := v.ValidateSignal(signal, 50000, 0.5)
	if !result.Valid {
		t.Errorf("expected valid signal, got errors: %v", result.Errors)
	}
}

func TestValidateSignal_ValidSellSignal(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.Balances[1] = map[string]float64{"USDT": 10000}
	bc.TotalBalances[1] = 10000

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

	result := v.ValidateSignal(signal, 3000, 0.5)
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

	result := v.ValidateSignal(signal, 100, 0.1)
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

	result := v.ValidateSignal(signal, 100, 0.1)
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

	result := v.ValidateSignal(signal, 100, 0.1)
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

	// 5% slippage, max 1%
	result := v.ValidateSignal(signal, 47500, 0.1)
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
	result := v.ValidateSignal(signal, 49500, 0.1)
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

	result := v.ValidateSignal(signal, 50000, 0.1)
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

	result := v.ValidateSignal(signal, 50000, 0.1)
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

	result := v.ValidateSignal(signal, 50000, 0.1)
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

	result := v.ValidateSignal(signal, 50000, 0.1)
	if result.Valid {
		t.Error("expected invalid signal for bad take-profit on SELL")
	}
}

func TestValidateSignal_InsufficientBalance(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.Balances[1] = map[string]float64{"USDT": 100}
	bc.TotalBalances[1] = 100

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

	result := v.ValidateSignal(signal, 50000, 1.0)
	if result.Valid {
		t.Error("expected invalid signal for insufficient balance")
	}
}

func TestValidateSignal_ExposureLimitExceeded(t *testing.T) {
	bc := NewMockBalanceChecker()
	bc.Balances[1] = map[string]float64{"USDT": 100000}
	bc.TotalBalances[1] = 100000

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

	result := v.ValidateSignal(signal, 50000, 0.10)
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

	result := v.ValidateSignal(signal, 50000, 0.1)
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

	result := v.ValidateSignal(signal, 50000, 0.1)
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
	result := v.ValidateSignal(signal, 49000, 0.1)
	if !result.Valid {
		t.Errorf("expected valid signal with custom max slippage, got errors: %v", result.Errors)
	}
}
