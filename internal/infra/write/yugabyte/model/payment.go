package model

import "time"

// PaymentIntentRow is the Yugabyte persistence model for payment intents.
type PaymentIntentRow struct {
	IntentID         string    `db:"payment_intent_id"`
	OrderID          string    `db:"order_id"`
	Provider         string    `db:"provider"`
	ProviderIntentID string    `db:"provider_intent_id"`
	CheckoutURL      string    `db:"checkout_url"`
	AmountCents      int64     `db:"amount_cents"`
	Currency         string    `db:"currency"`
	Status           string    `db:"status"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
