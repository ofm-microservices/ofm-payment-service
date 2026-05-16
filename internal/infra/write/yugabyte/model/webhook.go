package model

import "time"

// WebhookRow is the Yugabyte persistence model for deduplicated webhooks.
type WebhookRow struct {
	EventID   string    `db:"event_id"`
	Provider  string    `db:"provider"`
	OrderID   string    `db:"order_id"`
	IntentID  string    `db:"payment_intent_id"`
	Status    string    `db:"status"`
	Payload   string    `db:"payload_json"`
	CreatedAt time.Time `db:"created_at"`
}
