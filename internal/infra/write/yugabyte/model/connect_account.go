package model

import "time"

// ConnectAccountRow mirrors the Stripe Connect onboarding write model.
type ConnectAccountRow struct {
	UserID           string    `db:"user_id"`
	StripeAccountID  string    `db:"stripe_account_id"`
	Status           string    `db:"status"`
	DetailsSubmitted bool      `db:"details_submitted"`
	ChargesEnabled   bool      `db:"charges_enabled"`
	PayoutsEnabled   bool      `db:"payouts_enabled"`
	DisabledReason   string    `db:"disabled_reason"`
	OnboardingURL    string    `db:"onboarding_url"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
