package stripe

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	app "payment-service/internal/application"
	"payment-service/internal/domain"
)

type fakeGateway struct{ log logging.Logger }

// NewFake constructs a deterministic Stripe substitute for automated tests.
// It performs no network calls and returns completed Connect accounts so tests
// can exercise publish and order flows without browser-based onboarding.
func NewFake(log logging.Logger) (app.StripeConnectGateway, error) {
	if log == nil {
		return nil, ErrNilLogger
	}
	return &fakeGateway{log: log.With(logging.String("module", "stripe-fake-gateway"))}, nil
}

func (g *fakeGateway) CreateAccount(_ context.Context, cmd app.StartFreelancerOnboardingCommand) (*domain.ConnectAccount, error) {
	id := stableID("acct", cmd.UserID)
	now := time.Now().UTC()
	return &domain.ConnectAccount{UserID: cmd.UserID, StripeAccountID: id, Status: domain.ConnectStatusCompleted, DetailsSubmitted: true, ChargesEnabled: true, PayoutsEnabled: true, CreatedAt: now, UpdatedAt: now}, nil
}

func (g *fakeGateway) CreateOnboardingLink(_ context.Context, accountID, returnURL, _ string) (string, error) {
	return fmt.Sprintf("%s?fake_stripe_account=%s", firstNonEmpty(returnURL, "http://localhost/fake-stripe/return"), accountID), nil
}

func (g *fakeGateway) CreateTransfer(_ context.Context, cmd app.ReleaseFundsCommand, _ string) (string, error) {
	return stableID("tr", firstNonEmpty(cmd.IdempotencyKey, cmd.OrderID)), nil
}

func (g *fakeGateway) CreateRefund(_ context.Context, paymentIntentID string, amountCents int64, idempotencyKey string) (string, error) {
	if amountCents <= 0 {
		return "", nil
	}
	return stableID("re", firstNonEmpty(idempotencyKey, paymentIntentID)), nil
}

func stableID(prefix, value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return prefix + "_fake_" + hex.EncodeToString(sum[:])[:24]
}
