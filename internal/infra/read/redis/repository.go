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
	return nil
}
