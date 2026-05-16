package stripe

import (
	"context"
	"strings"
	"time"

	"payment-service/config"
	app "payment-service/internal/application"
	"payment-service/internal/domain"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	stripec "github.com/stripe/stripe-go/v85"
	account "github.com/stripe/stripe-go/v85/account"
	accountlink "github.com/stripe/stripe-go/v85/accountlink"
)

type gateway struct {
	cfg config.StripeConfig
	log logging.Logger
}

// New constructs the Stripe Connect adapter used by payment-service.
func New(cfg config.StripeConfig, log logging.Logger) (app.StripeConnectGateway, error) {
	if strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, ErrCreateAccount
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	stripec.Key = cfg.SecretKey
	return &gateway{cfg: cfg, log: log.With(logging.String("module", "stripe-connect-gateway"))}, nil
}

func (g *gateway) CreateAccount(ctx context.Context, cmd app.StartFreelancerOnboardingCommand) (*domain.ConnectAccount, error) {
	_ = ctx
	params := &stripec.AccountParams{
		Country:      stripec.String(firstNonEmpty(cmd.Country, g.cfg.ConnectCountry)),
		Type:         stripec.String(string(stripec.AccountTypeExpress)),
		Capabilities: &stripec.AccountCapabilitiesParams{CardPayments: &stripec.AccountCapabilitiesCardPaymentsParams{Requested: stripec.Bool(true)}, Transfers: &stripec.AccountCapabilitiesTransfersParams{Requested: stripec.Bool(true)}},
		Metadata: map[string]string{
			"user_id": cmd.UserID,
		},
	}
	params.SetIdempotencyKey("freelancer-onboarding-account:" + strings.TrimSpace(cmd.UserID))
	acct, err := account.New(params)
	if err != nil {
		g.log.Error("stripe api failed",
			logging.Operation("stripe.connect.create_account"),
			logging.String("user_id", cmd.UserID),
			logging.Err(err),
		)
		return nil, ErrCreateAccount
	}
	now := time.Now().UTC()
	return &domain.ConnectAccount{
		UserID:          cmd.UserID,
		StripeAccountID: acct.ID,
		Status:          domain.ConnectStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (g *gateway) CreateOnboardingLink(ctx context.Context, accountID, returnURL, refreshURL string) (string, error) {
	_ = ctx
	link, err := accountlink.New(&stripec.AccountLinkParams{
		Account:    stripec.String(accountID),
		ReturnURL:  stripec.String(firstNonEmpty(returnURL, g.cfg.ConnectReturnURL)),
		RefreshURL: stripec.String(firstNonEmpty(refreshURL, g.cfg.ConnectRefreshURL)),
		Type:       stripec.String(string(stripec.AccountLinkTypeAccountOnboarding)),
	})
	if err != nil {
		g.log.Error("stripe api failed",
			logging.Operation("stripe.connect.create_onboarding_link"),
			logging.String("account_id", accountID),
			logging.Err(err),
		)
		return "", ErrCreateOnboardingLink
	}
	return link.URL, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
