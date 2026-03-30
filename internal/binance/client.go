package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/khbm0110/copy-trading-platform/internal/metrics"
)

const (
	defaultBaseURL = "https://api.binance.com"
	apiOrderPath   = "/api/v3/order"
)

// OrderRequest represents a new order request to Binance.
type OrderRequest struct {
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Type          string  `json:"type"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	ClientOrderID string  `json:"newClientOrderId"`
	TimeInForce   string  `json:"timeInForce"`
}

// OrderResponse represents the Binance API response for an order.
type OrderResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       int64  `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	Status        string `json:"status"`
	ExecutedQty   string `json:"executedQty"`
	CumulativeQuoteQty string `json:"cummulativeQuoteQty"`
}

// QueryOrderResponse represents the Binance order query response.
type QueryOrderResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       int64  `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	Status        string `json:"status"`
	ExecutedQty   string `json:"executedQty"`
}

// APIError represents a Binance API error response.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("binance api error %d: %s", e.Code, e.Message)
}

// IsRateLimited returns true if this is a rate limit error.
func (e *APIError) IsRateLimited() bool {
	return e.Code == -1015
}

// IsRetriable returns true if this error should be retried.
func (e *APIError) IsRetriable() bool {
	return e.Code == -1015 || e.Code == -1001 || e.Code == -1000
}

// Client defines the interface for Binance API operations.
type Client interface {
	ExecuteOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error)
	QueryOrderStatus(ctx context.Context, symbol string, clientOrderID string) (*QueryOrderResponse, error)
	GetTickerPrice(ctx context.Context, symbol string) (float64, error)
	GetBalance(ctx context.Context, symbol string) (float64, error)
}

// RealClient implements Client with actual Binance API calls.
// Each instance is bound to a specific user's API credentials.
type RealClient struct {
	apiKey    string
	secretKey string
	baseURL   string
	httpClient *http.Client
	logger    *slog.Logger
	userID    int64
}

// ClientConfig holds configuration for creating a RealClient.
type ClientConfig struct {
	APIKey    string
	SecretKey string
	BaseURL   string
	UserID    int64
	Timeout   time.Duration
}

// NewRealClient creates a new per-user Binance API client.
func NewRealClient(cfg ClientConfig, logger *slog.Logger) *RealClient {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &RealClient{
		apiKey:    cfg.APIKey,
		secretKey: cfg.SecretKey,
		baseURL:   baseURL,
		httpClient: &http.Client{Timeout: timeout},
		logger:    logger.With(slog.Int64("user_id", cfg.UserID)),
		userID:    cfg.UserID,
	}
}

func (c *RealClient) ExecuteOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	start := time.Now()
	defer func() {
		metrics.BinanceAPIDuration.WithLabelValues("execute_order").Observe(time.Since(start).Seconds())
	}()

	c.logger.Info("executing order",
		slog.String("symbol", req.Symbol),
		slog.String("side", req.Side),
		slog.String("client_order_id", req.ClientOrderID),
		slog.Float64("quantity", req.Quantity),
		slog.Float64("price", req.Price),
	)

	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("side", req.Side)
	params.Set("type", "LIMIT")
	params.Set("timeInForce", "GTC")
	params.Set("quantity", strconv.FormatFloat(req.Quantity, 'f', 8, 64))
	params.Set("price", strconv.FormatFloat(req.Price, 'f', 8, 64))
	params.Set("newClientOrderId", req.ClientOrderID)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	signature := c.sign(params.Encode())
	params.Set("signature", signature)

	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, apiOrderPath, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		metrics.BinanceAPIErrors.WithLabelValues("network").Inc()
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			metrics.BinanceAPIErrors.WithLabelValues("unknown").Inc()
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}
		if apiErr.IsRateLimited() {
			metrics.BinanceAPIErrors.WithLabelValues("rate_limit").Inc()
		} else {
			metrics.BinanceAPIErrors.WithLabelValues("api_error").Inc()
		}
		return nil, &apiErr
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	c.logger.Info("order executed",
		slog.String("symbol", orderResp.Symbol),
		slog.Int64("binance_order_id", orderResp.OrderID),
		slog.String("status", orderResp.Status),
		slog.String("client_order_id", orderResp.ClientOrderID),
	)

	return &orderResp, nil
}

func (c *RealClient) QueryOrderStatus(ctx context.Context, symbol string, clientOrderID string) (*QueryOrderResponse, error) {
	start := time.Now()
	defer func() {
		metrics.BinanceAPIDuration.WithLabelValues("query_order").Observe(time.Since(start).Seconds())
	}()

	c.logger.Info("querying order status",
		slog.String("symbol", symbol),
		slog.String("client_order_id", clientOrderID),
	)

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("origClientOrderId", clientOrderID)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	signature := c.sign(params.Encode())
	params.Set("signature", signature)

	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, apiOrderPath, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		metrics.BinanceAPIErrors.WithLabelValues("network").Inc()
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}
		return nil, &apiErr
	}

	var queryResp QueryOrderResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	c.logger.Info("order status queried",
		slog.String("symbol", queryResp.Symbol),
		slog.Int64("binance_order_id", queryResp.OrderID),
		slog.String("status", queryResp.Status),
	)

	return &queryResp, nil
}

func (c *RealClient) GetTickerPrice(ctx context.Context, symbol string) (float64, error) {
	start := time.Now()
	defer func() {
		metrics.BinanceAPIDuration.WithLabelValues("get_ticker_price").Observe(time.Since(start).Seconds())
	}()

	params := url.Values{}
	params.Set("symbol", symbol)
	reqURL := fmt.Sprintf("%s/api/v3/ticker/price?%s", c.baseURL, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		metrics.BinanceAPIErrors.WithLabelValues("network").Inc()
		return 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var priceResp struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return 0, fmt.Errorf("unmarshal response: %w", err)
	}

	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("parse price: %w", err)
	}

	return price, nil
}

func (c *RealClient) GetBalance(ctx context.Context, symbol string) (float64, error) {
	start := time.Now()
	defer func() {
		metrics.BinanceAPIDuration.WithLabelValues("get_balance").Observe(time.Since(start).Seconds())
	}()

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")
	signature := c.sign(params.Encode())
	params.Set("signature", signature)

	reqURL := fmt.Sprintf("%s/api/v3/account?%s", c.baseURL, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		metrics.BinanceAPIErrors.WithLabelValues("network").Inc()
		return 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var accountResp struct {
		Balances []struct {
			Asset string `json:"asset"`
			Free  string `json:"free"`
		} `json:"balances"`
	}
	if err := json.Unmarshal(body, &accountResp); err != nil {
		return 0, fmt.Errorf("unmarshal response: %w", err)
	}

	for _, b := range accountResp.Balances {
		if b.Asset == symbol {
			free, err := strconv.ParseFloat(b.Free, 64)
			if err != nil {
				return 0, fmt.Errorf("parse free balance: %w", err)
			}
			return free, nil
		}
	}

	return 0, nil
}

// sign creates an HMAC-SHA256 signature for Binance API authentication.
func (c *RealClient) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}