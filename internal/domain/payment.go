package domain

import (
	"context"
	"time"
)

const (
	PaymentStatusPending        = "pending"
	PaymentStatusIntentCreated  = "intent_created"
	PaymentStatusRequiresAction = "requires_action"
	PaymentStatusCaptured       = "captured"
	PaymentStatusFailed         = "failed"
	PaymentStatusExpired        = "expired"
	PaymentStatusRefunded       = "refunded"
)

// PaymentIntent is the write-model entity owned by payment-service.
type PaymentIntent struct {
	IntentID         string
	OrderID          string
	Provider         string
	ProviderIntentID string
	CheckoutURL      string
	AmountCents      int64
	Currency         string
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

const (
	PaymentReleaseStatusPending  = "pending"
	PaymentReleaseStatusReleased = "released"
	PaymentReleaseStatusFailed   = "failed"
)

// PaymentRelease stores a payout transfer attempt to a seller.
type PaymentRelease struct {
	ReleaseID        string
	OrderID          string
	PaymentID        string
	SellerUserID     string
	AmountCents      int64
	Currency         string
	IdempotencyKey   string
	StripeTransferID string
	Status           string
	FailureReason    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PaymentIntentCommand is the command payload persisted for intent creation.
type PaymentIntentCommand struct {
	IntentID       string
	SagaID         string
	OrderID        string
	AmountCents    int64
	Currency       string
	Provider       string
	IdempotencyKey string
	RequestedAt    time.Time
}

// WebhookEvent stores deduplicated provider webhook metadata.
type WebhookEvent struct {
	EventID   string
	Provider  string
	OrderID   string
	IntentID  string
	Status    string
	Payload   string
	CreatedAt time.Time
}

const (
	// ConnectStatusPending indicates that the Stripe account exists but the
	// freelancer still needs to finish onboarding.
	ConnectStatusPending = "pending"
	// ConnectStatusCompleted indicates that Stripe has confirmed the account is
	// ready for payouts.
	ConnectStatusCompleted = "completed"
	// ConnectStatusDisabled indicates that Stripe disabled the account or
	// requires manual review.
	ConnectStatusDisabled = "disabled"
)

// ConnectAccount represents a freelancer Stripe Connect account stored by
// payment-service as the onboarding source of truth.
type ConnectAccount struct {
	UserID           string
	StripeAccountID  string
	Status           string
	DetailsSubmitted bool
	ChargesEnabled   bool
	PayoutsEnabled   bool
	DisabledReason   string
	OnboardingURL    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ConnectAccountRepository persists the freelancer Stripe Connect account
// write model.
type ConnectAccountRepository interface {
	Upsert(ctx context.Context, account ConnectAccount) (*ConnectAccount, error)
	GetByUserID(ctx context.Context, userID string) (*ConnectAccount, error)
	GetByStripeAccountID(ctx context.Context, accountID string) (*ConnectAccount, error)
}

// PaymentIntentRepository persists the payment intent write model.
type PaymentIntentRepository interface {
	Create(ctx context.Context, intent PaymentIntent) (*PaymentIntent, error)
	GetByID(ctx context.Context, intentID string) (*PaymentIntent, error)
	GetByOrderID(ctx context.Context, orderID string) (*PaymentIntent, error)
	UpdateWebhookPaymentIntent(ctx context.Context, orderID, providerIntentID, status string) (*PaymentIntent, error)
	UpdateStatus(ctx context.Context, intentID, status string) error
	UpdateCheckoutURL(ctx context.Context, intentID, checkoutURL string) error
}

// PaymentReleaseRepository persists seller payout releases.
type PaymentReleaseRepository interface {
	Create(ctx context.Context, release PaymentRelease) (*PaymentRelease, error)
	GetByOrderID(ctx context.Context, orderID string) (*PaymentRelease, error)
	UpdateStatus(ctx context.Context, releaseID, status, transferID, failureReason string) error
}

// PaymentIntentReadRepository persists the payment read model projection.
type PaymentIntentReadRepository interface {
	Upsert(ctx context.Context, intent *PaymentIntent) error
}

// WebhookRepository persists deduplicated webhook events.
type WebhookRepository interface {
	Create(ctx context.Context, evt WebhookEvent) (*WebhookEvent, error)
	Exists(ctx context.Context, eventID string) (bool, error)
}
