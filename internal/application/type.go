package application

import (
	"context"
	"payment-service/internal/domain"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
)

// Service owns payment intent creation and Stripe webhook normalization.
type Service interface {
	CreateIntent(ctx context.Context, cmd CreateIntentCommand) (*CreateIntentResult, error)
	HandleWebhook(ctx context.Context, evt WebhookCommand) error
	StartFreelancerOnboarding(ctx context.Context, cmd StartFreelancerOnboardingCommand) (*StartFreelancerOnboardingResult, error)
	GetConnectStatus(ctx context.Context, userID string) (*GetConnectStatusResult, error)
	HandleConnectWebhook(ctx context.Context, evt ConnectWebhookCommand) error
	ReleaseFunds(ctx context.Context, cmd ReleaseFundsCommand) (*ReleaseFundsResult, error)
	GetReleaseByOrderID(ctx context.Context, orderID string) (*GetReleaseByOrderResult, error)
}

// IntentRepository aliases the domain payment intent repository.
type IntentRepository = domain.PaymentIntentRepository

// WebhookRepository aliases the domain webhook repository.
type WebhookRepository = domain.WebhookRepository

// PaymentIntentReadRepository aliases the payment read-model repository.
type PaymentIntentReadRepository = domain.PaymentIntentReadRepository

// ConnectAccountRepository aliases the Stripe Connect onboarding repository.
type ConnectAccountRepository = domain.ConnectAccountRepository

// PaymentReleaseRepository aliases the seller payout release repository.
type PaymentReleaseRepository = domain.PaymentReleaseRepository

// StripeConnectGateway abstracts the Stripe Connect onboarding API calls.
type StripeConnectGateway interface {
	CreateAccount(ctx context.Context, cmd StartFreelancerOnboardingCommand) (*domain.ConnectAccount, error)
	CreateOnboardingLink(ctx context.Context, accountID, returnURL, refreshURL string) (string, error)
	CreateTransfer(ctx context.Context, cmd ReleaseFundsCommand, destinationAccountID string) (string, error)
}

// EventBroker abstracts the runtime NATS broker.
type EventBroker interface {
	Publish(ctx context.Context, subject string, payload []byte) error
	Subscribe(ctx context.Context, subject string, handler MessageHandler) error
	Close()
}

// MessageHandler processes one broker payload.
type MessageHandler func(ctx context.Context, subject string, payload []byte) error

// Logger aliases the shared structured logger.
type Logger = logging.Logger

// CreateIntentCommand starts a new payment intent.
type CreateIntentCommand struct {
	IntentID       string `json:"payment_intent_id"`
	SagaID         string `json:"saga_id"`
	OrderID        string `json:"order_id"`
	AmountCents    int64  `json:"amount_cents"`
	Currency       string `json:"currency"`
	Provider       string `json:"provider"`
	IdempotencyKey string `json:"idempotency_key"`
	RequestedAt    string `json:"requested_at"`
}

// CreateIntentResult reports the payment intent outcome.
type CreateIntentResult struct {
	IntentID         string `json:"payment_intent_id"`
	SagaID           string `json:"saga_id"`
	OrderID          string `json:"order_id"`
	ProviderIntentID string `json:"provider_intent_id,omitempty"`
	CheckoutURL      string `json:"checkout_url,omitempty"`
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
	OccurredAt       string `json:"occurred_at"`
}

// WebhookCommand is the normalized Stripe webhook input.
type WebhookCommand struct {
	EventID          string `json:"event_id"`
	Provider         string `json:"provider"`
	EventType        string `json:"event_type"`
	SagaID           string `json:"saga_id"`
	IntentID         string `json:"payment_intent_id"`
	ProviderIntentID string `json:"provider_intent_id"`
	OrderID          string `json:"order_id"`
	Status           string `json:"status"`
	PayloadJSON      string `json:"payload_json"`
	OccurredAt       string `json:"occurred_at"`
}

// StartFreelancerOnboardingCommand starts Stripe Connect onboarding for one
// authenticated freelancer.
type StartFreelancerOnboardingCommand struct {
	UserID         string `json:"user_id"`
	Country        string `json:"country"`
	ReturnURL      string `json:"return_url"`
	RefreshURL     string `json:"refresh_url"`
	IdempotencyKey string `json:"idempotency_key"`
	RequestedAt    string `json:"requested_at"`
}

// StartFreelancerOnboardingResult reports the Stripe Connect onboarding
// account and URL created for the user.
type StartFreelancerOnboardingResult struct {
	UserID           string `json:"user_id"`
	StripeAccountID  string `json:"stripe_account_id"`
	OnboardingURL    string `json:"onboarding_url,omitempty"`
	Status           string `json:"status"`
	DetailsSubmitted bool   `json:"details_submitted"`
	ChargesEnabled   bool   `json:"charges_enabled"`
	PayoutsEnabled   bool   `json:"payouts_enabled"`
	DisabledReason   string `json:"disabled_reason,omitempty"`
	OccurredAt       string `json:"occurred_at"`
}

// GetConnectStatusResult reports the current Stripe Connect onboarding status
// for one freelancer.
type GetConnectStatusResult struct {
	UserID          string `json:"user_id"`
	StripeAccountID string `json:"stripe_account_id"`
	Status          string `json:"status"`
	DisabledReason  string `json:"disabled_reason,omitempty"`
	OccurredAt      string `json:"occurred_at"`
}

// ConnectWebhookCommand is the normalized Stripe Connect webhook input.
type ConnectWebhookCommand struct {
	EventID          string `json:"event_id"`
	Provider         string `json:"provider"`
	EventType        string `json:"event_type"`
	UserID           string `json:"user_id,omitempty"`
	StripeAccountID  string `json:"stripe_account_id"`
	DetailsSubmitted bool   `json:"details_submitted"`
	ChargesEnabled   bool   `json:"charges_enabled"`
	PayoutsEnabled   bool   `json:"payouts_enabled"`
	DisabledReason   string `json:"disabled_reason,omitempty"`
	PayloadJSON      string `json:"payload_json"`
	OccurredAt       string `json:"occurred_at"`
}

// ReleaseFundsCommand requests a seller payout transfer for a captured payment.
type ReleaseFundsCommand struct {
	OrderID        string `json:"order_id"`
	PaymentID      string `json:"payment_id"`
	SellerUserID   string `json:"seller_user_id"`
	AmountCents    int64  `json:"amount_cents"`
	Currency       string `json:"currency"`
	IdempotencyKey string `json:"idempotency_key"`
	RequestedAt    string `json:"requested_at"`
}

// ReleaseFundsResult reports the payout transfer outcome.
type ReleaseFundsResult struct {
	OrderID          string `json:"order_id"`
	PaymentReleaseID string `json:"payment_release_id"`
	StripeTransferID string `json:"stripe_transfer_id"`
	Status           string `json:"status"`
	OccurredAt       string `json:"occurred_at"`
}

// GetReleaseByOrderResult reports a persisted payout release.
type GetReleaseByOrderResult struct {
	OrderID          string `json:"order_id"`
	PaymentReleaseID string `json:"payment_release_id"`
	PaymentID        string `json:"payment_intent_id"`
	SellerUserID     string `json:"seller_user_id"`
	AmountCents      int64  `json:"amount_cents"`
	Currency         string `json:"currency"`
	IdempotencyKey   string `json:"idempotency_key"`
	StripeTransferID string `json:"stripe_transfer_id"`
	Status           string `json:"status"`
	FailureReason    string `json:"failure_reason"`
	OccurredAt       string `json:"occurred_at"`
}
