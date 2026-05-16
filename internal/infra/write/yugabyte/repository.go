package yugabyte

import (
	"context"
	"time"

	"payment-service/internal/domain"
	"payment-service/internal/infra/write/yugabyte/model"

	"github.com/jmoiron/sqlx"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
)

type repo struct {
	db         *sqlx.DB
	translator DBErrorTranslator
	log        logging.Logger
}

// New constructs the Yugabyte-backed payment repository.
func New(db *sqlx.DB, translator DBErrorTranslator, log logging.Logger) (domain.PaymentIntentRepository, error) {
	if db == nil {
		return nil, ErrNilYugaByteDB
	}
	if translator == nil {
		return nil, ErrNilDBErrorTranslator
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &repo{db: db, translator: translator, log: log.With(logging.String("module", "yugabyte-repository"))}, nil
}

const createPaymentIntentQuery = `
INSERT INTO payment_intents (payment_intent_id, order_id, provider, provider_intent_id, checkout_url, amount_cents, currency, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING payment_intent_id, order_id, provider, provider_intent_id, checkout_url, amount_cents, currency, status, created_at, updated_at`

const getPaymentIntentByIDQuery = `
SELECT payment_intent_id, order_id, provider, provider_intent_id, checkout_url, amount_cents, currency, status, created_at, updated_at
FROM payment_intents WHERE payment_intent_id = $1`

const updatePaymentIntentStatusQuery = `UPDATE payment_intents SET status = $2, updated_at = NOW() WHERE payment_intent_id = $1`

const updatePaymentIntentCheckoutURLQuery = `UPDATE payment_intents SET checkout_url = $2, updated_at = NOW() WHERE payment_intent_id = $1`
const createWebhookEventQuery = `
INSERT INTO payment_webhook_events (event_id, provider, order_id, payment_intent_id, status, payload_json, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING event_id, provider, order_id, payment_intent_id, status, payload_json, created_at`
const existsWebhookEventQuery = `SELECT EXISTS (SELECT 1 FROM payment_webhook_events WHERE event_id = $1)`

func (r *repo) Create(ctx context.Context, intent domain.PaymentIntent) (*domain.PaymentIntent, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "create", "payment_intents", status, time.Since(started))
	}()

	var row model.PaymentIntentRow
	if err := r.db.QueryRowContext(ctx, createPaymentIntentQuery,
		intent.IntentID, intent.OrderID, intent.Provider, intent.ProviderIntentID, intent.CheckoutURL,
		intent.AmountCents, intent.Currency, intent.Status, intent.CreatedAt, intent.UpdatedAt,
	).Scan(
		&row.IntentID, &row.OrderID, &row.Provider, &row.ProviderIntentID, &row.CheckoutURL,
		&row.AmountCents, &row.Currency, &row.Status, &row.CreatedAt, &row.UpdatedAt,
	); err != nil {
		status = "error"
		r.log.Error("create payment intent failed",
			logging.Operation("db.payment.create"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("payment_intent_id", intent.IntentID),
			logging.Err(err),
		)
		return nil, r.translator.TranslateCreatePaymentIntentError(err)
	}
	return &domain.PaymentIntent{
		IntentID: row.IntentID, OrderID: row.OrderID, Provider: row.Provider,
		ProviderIntentID: row.ProviderIntentID, CheckoutURL: row.CheckoutURL,
		AmountCents: row.AmountCents, Currency: row.Currency, Status: row.Status,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *repo) GetByID(ctx context.Context, intentID string) (*domain.PaymentIntent, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "get_by_id", "payment_intents", status, time.Since(started))
	}()

	var row model.PaymentIntentRow
	if err := r.db.QueryRowContext(ctx, getPaymentIntentByIDQuery, intentID).Scan(
		&row.IntentID, &row.OrderID, &row.Provider, &row.ProviderIntentID, &row.CheckoutURL,
		&row.AmountCents, &row.Currency, &row.Status, &row.CreatedAt, &row.UpdatedAt,
	); err != nil {
		status = "error"
		r.log.Error("get payment intent failed",
			logging.Operation("db.payment.get_by_id"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("payment_intent_id", intentID),
			logging.Err(err),
		)
		return nil, r.translator.TranslateFindPaymentIntentError(err)
	}
	return &domain.PaymentIntent{
		IntentID: row.IntentID, OrderID: row.OrderID, Provider: row.Provider,
		ProviderIntentID: row.ProviderIntentID, CheckoutURL: row.CheckoutURL,
		AmountCents: row.AmountCents, Currency: row.Currency, Status: row.Status,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *repo) UpdateStatus(ctx context.Context, intentID, statusValue string) error {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "update_status", "payment_intents", status, time.Since(started))
	}()

	if _, err := r.db.ExecContext(ctx, updatePaymentIntentStatusQuery, intentID, statusValue); err != nil {
		status = "error"
		r.log.Error("update payment intent status failed",
			logging.Operation("db.payment.update_status"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("payment_intent_id", intentID),
			logging.Err(err),
		)
		return r.translator.TranslateUpdatePaymentIntentError(err)
	}
	return nil
}

func (r *repo) UpdateCheckoutURL(ctx context.Context, intentID, checkoutURL string) error {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "update_checkout_url", "payment_intents", status, time.Since(started))
	}()

	if _, err := r.db.ExecContext(ctx, updatePaymentIntentCheckoutURLQuery, intentID, checkoutURL); err != nil {
		status = "error"
		r.log.Error("update payment intent checkout url failed",
			logging.Operation("db.payment.update_checkout_url"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("payment_intent_id", intentID),
			logging.Err(err),
		)
		return r.translator.TranslateUpdatePaymentIntentError(err)
	}
	return nil
}

// WebhookRepo persists deduplicated webhook events.
type WebhookRepo struct {
	db         *sqlx.DB
	translator DBErrorTranslator
	log        logging.Logger
}

// NewWebhookRepo constructs the Yugabyte-backed webhook repository.
func NewWebhookRepo(db *sqlx.DB, translator DBErrorTranslator, log logging.Logger) (domain.WebhookRepository, error) {
	if db == nil {
		return nil, ErrNilYugaByteDB
	}
	if translator == nil {
		return nil, ErrNilDBErrorTranslator
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &WebhookRepo{db: db, translator: translator, log: log.With(logging.String("module", "yugabyte-webhook-repository"))}, nil
}

func (r *WebhookRepo) Create(ctx context.Context, evt domain.WebhookEvent) (*domain.WebhookEvent, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "create", "payment_webhook_events", status, time.Since(started))
	}()

	var row model.WebhookRow
	if err := r.db.QueryRowContext(ctx, createWebhookEventQuery,
		evt.EventID, evt.Provider, evt.OrderID, evt.IntentID, evt.Status, evt.Payload, evt.CreatedAt,
	).Scan(&row.EventID, &row.Provider, &row.OrderID, &row.IntentID, &row.Status, &row.Payload, &row.CreatedAt); err != nil {
		status = "error"
		return nil, r.translator.TranslateCreatePaymentIntentError(err)
	}
	return &domain.WebhookEvent{
		EventID: row.EventID, Provider: row.Provider, OrderID: row.OrderID,
		IntentID: row.IntentID, Status: row.Status, Payload: row.Payload, CreatedAt: row.CreatedAt,
	}, nil
}

func (r *WebhookRepo) Exists(ctx context.Context, eventID string) (bool, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "exists", "payment_webhook_events", status, time.Since(started))
	}()

	var exists bool
	if err := r.db.QueryRowContext(ctx, existsWebhookEventQuery, eventID).Scan(&exists); err != nil {
		status = "error"
		return false, r.translator.TranslateFindPaymentIntentError(err)
	}
	return exists, nil
}
