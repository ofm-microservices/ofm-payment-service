package model

// PaymentIntentCache is the Redis projection model for payment intents.
type PaymentIntentCache struct {
	IntentID    string `json:"payment_intent_id"`
	OrderID     string `json:"order_id"`
	Provider    string `json:"provider"`
	CheckoutURL string `json:"checkout_url"`
	Status      string `json:"status"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
