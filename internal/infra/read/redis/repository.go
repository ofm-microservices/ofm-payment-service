package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"payment-service/internal/domain"
	"payment-service/internal/infra/read/redis/model"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	"github.com/redis/go-redis/v9"
)

type repo struct {
	rdb *redis.Client
	log logging.Logger
}

var paymentByOrderScript = redis.NewScript(`
local order_key = KEYS[1]
local payment_key_prefix = KEYS[2]
local payment_id = redis.call("GET", order_key)
if not payment_id then
  return {err = "payment intent not found"}
end
local payload = redis.call("GET", payment_key_prefix .. payment_id)
if not payload then
  return {err = "payment intent not found"}
end
return {payment_id, payload}
`)

// New constructs the Redis-backed payment read repository.
func New(rdb *redis.Client, log logging.Logger) (domain.PaymentIntentReadRepository, error) {
	if rdb == nil {
		return nil, ErrNilRedisClient
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &repo{rdb: rdb, log: log.With(logging.String("module", "redis-repository"))}, nil
}

// PaymentIntentKey builds the Redis key used for payment-intent projections.
func PaymentIntentKey(intentID string) string { return fmt.Sprintf("payment_intent:%s", intentID) }

// PaymentOrderKey builds the Redis key used to index payments by order ID.
func PaymentOrderKey(orderID string) string { return fmt.Sprintf("order:%s", orderID) }

func (r *repo) Upsert(ctx context.Context, intent *domain.PaymentIntent) error {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("set", "payment_intent", status, time.Since(started)) }()
	if intent == nil {
		return ErrNilPaymentIntent
	}
	cache := model.PaymentIntentCache{
		IntentID: intent.IntentID, OrderID: intent.OrderID, Provider: intent.Provider,
		CheckoutURL: intent.CheckoutURL, Status: intent.Status, AmountCents: intent.AmountCents,
		Currency: intent.Currency, CreatedAt: intent.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: intent.UpdatedAt.Format(time.RFC3339Nano),
	}
	payload, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	if err := r.rdb.Set(ctx, PaymentIntentKey(intent.IntentID), payload, 0).Err(); err != nil {
		status = "error"
		r.log.Error("upsert payment intent cache failed", logging.Operation("redis.payment.upsert"), logging.DurationMS(time.Since(started)), logging.Err(err))
		return err
	}
	if err := r.rdb.Set(ctx, PaymentOrderKey(intent.OrderID), intent.IntentID, 0).Err(); err != nil {
		status = "error"
		r.log.Error("upsert payment order index failed", logging.Operation("redis.payment.upsert_order_index"), logging.String("order_id", intent.OrderID), logging.String("payment_intent_id", intent.IntentID), logging.DurationMS(time.Since(started)), logging.Err(err))
		return err
	}
	return nil
}

func (r *repo) GetByOrderID(ctx context.Context, orderID string) (*domain.PaymentIntent, error) {
	started := time.Now()
	status := "success"
	defer func() { metrics.Global().ObserveRedis("get", "payment_intent_by_order", status, time.Since(started)) }()
	if orderID == "" {
		return nil, domain.ErrInvalidID
	}
	raw, err := paymentByOrderScript.Run(ctx, r.rdb, []string{PaymentOrderKey(orderID), "payment_intent:"}).Result()
	if err != nil {
		status = "error"
		r.log.Error("get payment intent by order cache failed", logging.Operation("redis.payment.get_by_order_id"), logging.String("order_id", orderID), logging.DurationMS(time.Since(started)), logging.Err(err))
		return nil, domain.ErrIntentNotFound
	}
	arr, ok := raw.([]interface{})
	if !ok || len(arr) != 2 {
		status = "error"
		return nil, domain.ErrIntentNotFound
	}
	payload, ok := arr[1].(string)
	if !ok {
		status = "error"
		return nil, domain.ErrIntentNotFound
	}
	var cache model.PaymentIntentCache
	if err := json.Unmarshal([]byte(payload), &cache); err != nil {
		status = "error"
		r.log.Error("unmarshal payment intent cache failed", logging.Operation("redis.payment.get_by_order_id.unmarshal"), logging.String("order_id", orderID), logging.DurationMS(time.Since(started)), logging.Err(err))
		return nil, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, cache.CreatedAt)
	if err != nil {
		status = "error"
		return nil, err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, cache.UpdatedAt)
	if err != nil {
		status = "error"
		return nil, err
	}
	return &domain.PaymentIntent{
		IntentID:    cache.IntentID,
		OrderID:     cache.OrderID,
		Provider:    cache.Provider,
		CheckoutURL: cache.CheckoutURL,
		Status:      cache.Status,
		AmountCents: cache.AmountCents,
		Currency:    cache.Currency,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
