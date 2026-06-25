package model

import "time"

// PaymentReleaseRow mirrors the seller payout transfer persistence model.
type PaymentReleaseRow struct {
	ReleaseID             string    `db:"payment_release_id"`
	OrderID               string    `db:"order_id"`
	PaymentID             string    `db:"payment_intent_id"`
	SellerUserID          string    `db:"seller_user_id"`
	AmountCents           int64     `db:"amount_cents"`
	Currency              string    `db:"currency"`
	FreelancerPercentage  int32     `db:"freelancer_percentage"`
	CustomerPercentage    int32     `db:"customer_percentage"`
	FreelancerAmountCents int64     `db:"freelancer_amount_cents"`
	CustomerAmountCents   int64     `db:"customer_amount_cents"`
	IdempotencyKey        string    `db:"idempotency_key"`
	StripeTransferID      string    `db:"stripe_transfer_id"`
	StripeRefundID        string    `db:"stripe_refund_id"`
	Status                string    `db:"status"`
	FailureReason         string    `db:"failure_reason"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}
