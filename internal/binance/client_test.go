package binance

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRealClient_ExecuteOrder_Success(t *testing.T) {
	expectedResp := OrderResponse{
		Symbol:        "BTCUSDT",
		OrderID:       12345,
		ClientOrderID: "test-order-1",
		Status:        "FILLED",
		ExecutedQty:   "0.10000000",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("X-MBX-APIKEY") != "test-api-key" {
			t.Error("missing or incorrect API key header")
		}

		// Verify query parameters
		q := r.URL.Query()
		if q.Get("symbol") != "BTCUSDT" {
			t.Errorf("expected symbol BTCUSDT, got %s", q.Get("symbol"))
		}
		if q.Get("side") != "BUY" {
			t.Errorf("expected side BUY, got %s", q.Get("side"))
		}
		if q.Get("signature") == "" {
			t.Error("missing signature")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewRealClient(ClientConfig{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
		BaseURL:   server.URL,
		UserID:    1,
	}, testLogger())

	resp, err := client.ExecuteOrder(context.Background(), OrderRequest{
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      0.1,
		Price:         50000,
		ClientOrderID: "test-order-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OrderID != 12345 {
		t.Errorf("expected order ID 12345, got %d", resp.OrderID)
	}
	if resp.Status != "FILLED" {
		t.Errorf("expected FILLED, got %s", resp.Status)
	}
	if resp.ClientOrderID != "test-order-1" {
		t.Errorf("expected client order ID test-order-1, got %s", resp.ClientOrderID)
	}
}

func TestRealClient_ExecuteOrder_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{Code: -1121, Message: "Invalid symbol."})
	}))
	defer server.Close()

	client := NewRealClient(ClientConfig{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
		BaseURL:   server.URL,
		UserID:    1,
	}, testLogger())

	_, err := client.ExecuteOrder(context.Background(), OrderRequest{
		Symbol:        "INVALID",
		Side:          "BUY",
		Quantity:      1,
		Price:         100,
		ClientOrderID: "test-order-2",
	})

	if err == nil {
		t.Fatal("expected error for invalid symbol")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != -1121 {
		t.Errorf("expected error code -1121, got %d", apiErr.Code)
	}
}

func TestRealClient_ExecuteOrder_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(APIError{Code: -1015, Message: "Too many requests."})
	}))
	defer server.Close()

	client := NewRealClient(ClientConfig{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
		BaseURL:   server.URL,
		UserID:    1,
	}, testLogger())

	_, err := client.ExecuteOrder(context.Background(), OrderRequest{
		Symbol:        "BTCUSDT",
		Side:          "BUY",
		Quantity:      1,
		Price:         50000,
		ClientOrderID: "test-order-3",
	})

	if err == nil {
		t.Fatal("expected error for rate limit")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if !apiErr.IsRateLimited() {
		t.Error("expected rate limited error")
	}
	if !apiErr.IsRetriable() {
		t.Error("rate limit errors should be retriable")
	}
}

func TestRealClient_QueryOrderStatus_Success(t *testing.T) {
	expectedResp := QueryOrderResponse{
		Symbol:        "BTCUSDT",
		OrderID:       12345,
		ClientOrderID: "test-order-1",
		Status:        "FILLED",
		ExecutedQty:   "0.10000000",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-MBX-APIKEY") != "test-api-key" {
			t.Error("missing or incorrect API key header")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewRealClient(ClientConfig{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
		BaseURL:   server.URL,
		UserID:    1,
	}, testLogger())

	resp, err := client.QueryOrderStatus(context.Background(), "BTCUSDT", "test-order-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "FILLED" {
		t.Errorf("expected FILLED, got %s", resp.Status)
	}
}

func TestRealClient_QueryOrderStatus_OrderNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{Code: -2013, Message: "Order does not exist."})
	}))
	defer server.Close()

	client := NewRealClient(ClientConfig{
		APIKey:    "test-api-key",
		SecretKey: "test-secret-key",
		BaseURL:   server.URL,
		UserID:    1,
	}, testLogger())

	_, err := client.QueryOrderStatus(context.Background(), "BTCUSDT", "nonexistent")

	if err == nil {
		t.Fatal("expected error for nonexistent order")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != -2013 {
		t.Errorf("expected error code -2013, got %d", apiErr.Code)
	}
}

func TestRealClient_Sign(t *testing.T) {
	client := NewRealClient(ClientConfig{
		APIKey:    "test-key",
		SecretKey: "NhqPtmdSJYdKjVHjA7PZj4Mge3R5YNiP1e3UZjInClVN65XAbvqqM6A7H5fATj0j",
		UserID:    1,
	}, testLogger())

	// Test deterministic signing
	sig1 := client.sign("symbol=BTCUSDT&side=BUY")
	sig2 := client.sign("symbol=BTCUSDT&side=BUY")

	if sig1 != sig2 {
		t.Error("expected deterministic signatures")
	}
	if sig1 == "" {
		t.Error("expected non-empty signature")
	}
}

func TestAPIError_IsRetriable(t *testing.T) {
	tests := []struct {
		code      int
		retriable bool
	}{
		{-1015, true},  // rate limit
		{-1001, true},  // internal error
		{-1000, true},  // unknown
		{-1121, false}, // invalid symbol
		{-2010, false}, // insufficient balance
		{-2013, false}, // order not found
	}

	for _, tt := range tests {
		apiErr := &APIError{Code: tt.code}
		if apiErr.IsRetriable() != tt.retriable {
			t.Errorf("code %d: expected retriable=%v, got %v", tt.code, tt.retriable, apiErr.IsRetriable())
		}
	}
}
