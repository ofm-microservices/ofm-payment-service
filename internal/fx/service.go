package appfx

import (
	"payment-service/config"
	"payment-service/internal/application"
	"payment-service/internal/domain"
	stripeinfra "payment-service/internal/infra/stripe"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"go.uber.org/fx"
)

// ServiceModule provides the payment application service.
var ServiceModule = fx.Options(
	fx.Provide(ProvideStripeConnectGateway, ProvidePaymentService),
)

// ProvidePaymentService constructs the payment application service.
func ProvidePaymentService(writeRepo domain.PaymentIntentRepository, webhooks domain.WebhookRepository, read domain.PaymentIntentReadRepository, accounts domain.ConnectAccountRepository, releases domain.PaymentReleaseRepository, broker application.EventBroker, stripe application.StripeConnectGateway, cfg *config.Config, lg logging.Logger) (application.Service, error) {
	return application.New(writeRepo, webhooks, read, accounts, releases, broker, stripe, cfg.Stripe.SkipTransfers, lg)
}

// ProvideStripeConnectGateway constructs the Stripe Connect adapter.
func ProvideStripeConnectGateway(cfg *config.Config, lg logging.Logger) (application.StripeConnectGateway, error) {
	return stripeinfra.New(cfg.Stripe, lg)
}
