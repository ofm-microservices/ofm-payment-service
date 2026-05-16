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

type connectAccountRepo struct {
	db         *sqlx.DB
	translator DBErrorTranslator
	log        logging.Logger
}

// NewConnectAccountRepo constructs the Yugabyte-backed Stripe Connect onboarding repository.
func NewConnectAccountRepo(db *sqlx.DB, translator DBErrorTranslator, log logging.Logger) (domain.ConnectAccountRepository, error) {
	if db == nil {
		return nil, ErrNilYugaByteDB
	}
	if translator == nil {
		return nil, ErrNilDBErrorTranslator
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &connectAccountRepo{db: db, translator: translator, log: log.With(logging.String("module", "yugabyte-connect-account-repository"))}, nil
}

const upsertConnectAccountQuery = `
INSERT INTO connect_accounts (
  user_id, stripe_account_id, status, details_submitted, charges_enabled,
  payouts_enabled, disabled_reason, onboarding_url, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5,
  $6, $7, $8, $9, $10
)
ON CONFLICT (user_id) DO UPDATE SET
  stripe_account_id = EXCLUDED.stripe_account_id,
  status = CASE
    WHEN connect_accounts.status = 'completed' OR EXCLUDED.status = 'completed' THEN 'completed'
    WHEN connect_accounts.status = 'disabled' OR EXCLUDED.status = 'disabled' THEN 'disabled'
    ELSE 'pending'
  END,
  details_submitted = connect_accounts.details_submitted OR EXCLUDED.details_submitted,
  charges_enabled = connect_accounts.charges_enabled OR EXCLUDED.charges_enabled,
  payouts_enabled = connect_accounts.payouts_enabled OR EXCLUDED.payouts_enabled,
  disabled_reason = CASE
    WHEN connect_accounts.status = 'completed' THEN connect_accounts.disabled_reason
    WHEN EXCLUDED.status = 'completed' THEN EXCLUDED.disabled_reason
    WHEN connect_accounts.status = 'disabled' THEN connect_accounts.disabled_reason
    WHEN EXCLUDED.status = 'disabled' THEN EXCLUDED.disabled_reason
    ELSE EXCLUDED.disabled_reason
  END,
  onboarding_url = CASE
    WHEN connect_accounts.status IN ('completed', 'disabled') THEN connect_accounts.onboarding_url
    ELSE EXCLUDED.onboarding_url
  END,
  updated_at = EXCLUDED.updated_at
RETURNING user_id, stripe_account_id, status, details_submitted, charges_enabled, payouts_enabled, disabled_reason, onboarding_url, created_at, updated_at`

const getConnectAccountByUserIDQuery = `
SELECT user_id, stripe_account_id, status, details_submitted, charges_enabled, payouts_enabled, disabled_reason, onboarding_url, created_at, updated_at
FROM connect_accounts WHERE user_id = $1`

const getConnectAccountByStripeAccountIDQuery = `
SELECT user_id, stripe_account_id, status, details_submitted, charges_enabled, payouts_enabled, disabled_reason, onboarding_url, created_at, updated_at
FROM connect_accounts WHERE stripe_account_id = $1`

func (r *connectAccountRepo) Upsert(ctx context.Context, account domain.ConnectAccount) (*domain.ConnectAccount, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "upsert", "connect_accounts", status, time.Since(started))
	}()

	row := model.ConnectAccountRow{
		UserID:           account.UserID,
		StripeAccountID:  account.StripeAccountID,
		Status:           account.Status,
		DetailsSubmitted: account.DetailsSubmitted,
		ChargesEnabled:   account.ChargesEnabled,
		PayoutsEnabled:   account.PayoutsEnabled,
		DisabledReason:   account.DisabledReason,
		OnboardingURL:    account.OnboardingURL,
		CreatedAt:        account.CreatedAt,
		UpdatedAt:        account.UpdatedAt,
	}
	var persisted model.ConnectAccountRow
	if err := r.db.QueryRowxContext(ctx, upsertConnectAccountQuery,
		row.UserID, row.StripeAccountID, row.Status, row.DetailsSubmitted, row.ChargesEnabled,
		row.PayoutsEnabled, row.DisabledReason, row.OnboardingURL, row.CreatedAt, row.UpdatedAt,
	).StructScan(&persisted); err != nil {
		status = "error"
		r.log.Error("upsert connect account failed",
			logging.Operation("db.connect_account.upsert"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
		logging.String("user_id", account.UserID),
			logging.Err(err),
		)
		return nil, r.translator.TranslateCreatePaymentIntentError(err)
	}
	return &domain.ConnectAccount{
		UserID:           persisted.UserID,
		StripeAccountID:  persisted.StripeAccountID,
		Status:           persisted.Status,
		DetailsSubmitted: persisted.DetailsSubmitted,
		ChargesEnabled:   persisted.ChargesEnabled,
		PayoutsEnabled:   persisted.PayoutsEnabled,
		DisabledReason:   persisted.DisabledReason,
		OnboardingURL:    persisted.OnboardingURL,
		CreatedAt:        persisted.CreatedAt,
		UpdatedAt:        persisted.UpdatedAt,
	}, nil
}

func (r *connectAccountRepo) GetByUserID(ctx context.Context, userID string) (*domain.ConnectAccount, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "get_by_user_id", "connect_accounts", status, time.Since(started))
	}()

	var row model.ConnectAccountRow
	if err := r.db.QueryRowContext(ctx, getConnectAccountByUserIDQuery, userID).Scan(
		&row.UserID, &row.StripeAccountID, &row.Status, &row.DetailsSubmitted, &row.ChargesEnabled, &row.PayoutsEnabled, &row.DisabledReason, &row.OnboardingURL, &row.CreatedAt, &row.UpdatedAt,
	); err != nil {
		status = "error"
		return nil, domain.ErrConnectAccountNotFound
	}
	return &domain.ConnectAccount{
		UserID:           row.UserID,
		StripeAccountID:  row.StripeAccountID,
		Status:           row.Status,
		DetailsSubmitted: row.DetailsSubmitted,
		ChargesEnabled:   row.ChargesEnabled,
		PayoutsEnabled:   row.PayoutsEnabled,
		DisabledReason:   row.DisabledReason,
		OnboardingURL:    row.OnboardingURL,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}

func (r *connectAccountRepo) GetByStripeAccountID(ctx context.Context, accountID string) (*domain.ConnectAccount, error) {
	started := time.Now()
	status := "success"
	defer func() {
		metrics.Global().ObserveDB("yugabyte", "get_by_stripe_account_id", "connect_accounts", status, time.Since(started))
	}()

	var row model.ConnectAccountRow
	if err := r.db.QueryRowContext(ctx, getConnectAccountByStripeAccountIDQuery, accountID).Scan(
		&row.UserID, &row.StripeAccountID, &row.Status, &row.DetailsSubmitted, &row.ChargesEnabled, &row.PayoutsEnabled, &row.DisabledReason, &row.OnboardingURL, &row.CreatedAt, &row.UpdatedAt,
	); err != nil {
		status = "error"
		return nil, domain.ErrConnectAccountNotFound
	}
	return &domain.ConnectAccount{
		UserID:           row.UserID,
		StripeAccountID:  row.StripeAccountID,
		Status:           row.Status,
		DetailsSubmitted: row.DetailsSubmitted,
		ChargesEnabled:   row.ChargesEnabled,
		PayoutsEnabled:   row.PayoutsEnabled,
		DisabledReason:   row.DisabledReason,
		OnboardingURL:    row.OnboardingURL,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}
