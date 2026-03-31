package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// TradeSignal represents an incoming trade signal from a professional trader.
type TradeSignal struct {
	SignalID      string  `json:"signal_id"`
	UserID        int64   `json:"user_id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	StopLoss      float64 `json:"stop_loss,omitempty"`
	TakeProfit    float64 `json:"take_profit,omitempty"`
	MaxSlippage   float64 `json:"max_slippage,omitempty"`
	ClientOrderID string  `json:"client_order_id"`
}

// EventBus provides publish/subscribe functionality using Redis Streams.
type EventBus struct {
	client        *redis.Client
	logger        *slog.Logger
	consumerGroup string
	consumerName  string
}

// Config holds EventBus configuration.
// Uses *redis.Options for Redis connection settings.
type Config struct {
	RedisOptions  *redis.Options
	ConsumerGroup string
	ConsumerName  string
}

// New creates a new EventBus.
// redisOpts should be obtained from config.GetRedisOptions().
func New(cfg Config, logger *slog.Logger) (*EventBus, error) {
	if cfg.RedisOptions == nil {
		return nil, fmt.Errorf("redis options are required")
	}

	client := redis.NewClient(cfg.RedisOptions)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &EventBus{
		client:        client,
		logger:        logger,
		consumerGroup: cfg.ConsumerGroup,
		consumerName:  cfg.ConsumerName,
	}, nil
}

// EnsureStream creates the stream and consumer group if they don't exist.
func (eb *EventBus) EnsureStream(ctx context.Context, stream string) error {
	err := eb.client.XGroupCreateMkStream(ctx, stream, eb.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}
	return nil
}

// Publish publishes a trade signal to the given stream.
func (eb *EventBus) Publish(ctx context.Context, stream string, signal TradeSignal) (string, error) {
	data, err := json.Marshal(signal)
	if err != nil {
		return "", fmt.Errorf("marshal signal: %w", err)
	}

	id, err := eb.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("xadd: %w", err)
	}

	eb.logger.Info("published signal",
		slog.String("stream", stream),
		slog.String("message_id", id),
		slog.String("signal_id", signal.SignalID),
		slog.Int64("user_id", signal.UserID),
	)
	return id, nil
}

// Message wraps a Redis Stream message with parsed signal data.
type Message struct {
	ID     string
	Signal TradeSignal
	Raw    redis.XMessage
}

// Subscribe reads messages from the stream using consumer groups.
// It blocks until context is canceled. handler is called for each message.
// Messages are acknowledged after successful processing.
func (eb *EventBus) Subscribe(ctx context.Context, stream string, handler func(ctx context.Context, msg Message) error) error {
	eb.logger.Info("subscribing to stream",
		slog.String("stream", stream),
		slog.String("group", eb.consumerGroup),
		slog.String("consumer", eb.consumerName),
	)

	// First, process any pending messages (claimed but not ACKed)
	if err := eb.processPending(ctx, stream, handler); err != nil {
		eb.logger.Warn("error processing pending messages", slog.String("error", err.Error()))
	}

	// Then read new messages
	for {
		select {
		case <-ctx.Done():
			eb.logger.Info("subscriber shutting down", slog.String("stream", stream))
			return ctx.Err()
		default:
		}

		streams, err := eb.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    eb.consumerGroup,
			Consumer: eb.consumerName,
			Streams:  []string{stream, ">"},
			Count:    10,
			Block:    2 * time.Second,
		}).Result()
		if err != nil {
			if err == redis.Nil || ctx.Err() != nil {
				continue
			}
			eb.logger.Error("xreadgroup error", slog.String("error", err.Error()))
			time.Sleep(time.Second)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				if err := eb.processMessage(ctx, stream, msg, handler); err != nil {
					eb.logger.Error("message processing failed",
						slog.String("message_id", msg.ID),
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}
}

func (eb *EventBus) processPending(ctx context.Context, stream string, handler func(ctx context.Context, msg Message) error) error {
	pending, err := eb.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   stream,
		Group:    eb.consumerGroup,
		Start:    "-",
		End:      "+",
		Count:    100,
		Consumer: eb.consumerName,
	}).Result()
	if err != nil {
		return fmt.Errorf("xpending: %w", err)
	}

	for _, p := range pending {
		msgs, err := eb.client.XRangeN(ctx, stream, p.ID, p.ID, 1).Result()
		if err != nil || len(msgs) == 0 {
			continue
		}
		if err := eb.processMessage(ctx, stream, msgs[0], handler); err != nil {
			eb.logger.Warn("pending message processing failed",
				slog.String("message_id", p.ID),
				slog.String("error", err.Error()),
			)
		}
	}
	return nil
}

func (eb *EventBus) processMessage(ctx context.Context, stream string, raw redis.XMessage, handler func(ctx context.Context, msg Message) error) error {
	data, ok := raw.Values["data"].(string)
	if !ok {
		eb.logger.Warn("invalid message format, acknowledging", slog.String("message_id", raw.ID))
		return eb.client.XAck(ctx, stream, eb.consumerGroup, raw.ID).Err()
	}

	var signal TradeSignal
	if err := json.Unmarshal([]byte(data), &signal); err != nil {
		eb.logger.Warn("invalid signal JSON, acknowledging",
			slog.String("message_id", raw.ID),
			slog.String("error", err.Error()),
		)
		return eb.client.XAck(ctx, stream, eb.consumerGroup, raw.ID).Err()
	}

	msg := Message{
		ID:     raw.ID,
		Signal: signal,
		Raw:    raw,
	}

	if err := handler(ctx, msg); err != nil {
		return err
	}

	// Acknowledge on success
	return eb.client.XAck(ctx, stream, eb.consumerGroup, raw.ID).Err()
}

// RequeueWithDelay re-publishes a signal after a delay (for rate-limited messages).
func (eb *EventBus) RequeueWithDelay(ctx context.Context, stream string, signal TradeSignal, delay time.Duration) error {
	time.AfterFunc(delay, func() {
		requeueCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := eb.Publish(requeueCtx, stream, signal); err != nil {
			eb.logger.Error("requeue failed",
				slog.String("signal_id", signal.SignalID),
				slog.String("error", err.Error()),
			)
		}
	})
	return nil
}

// Close gracefully shuts down the EventBus.
func (eb *EventBus) Close() error {
	return eb.client.Close()
}
