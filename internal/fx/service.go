package appfx

import (
	"payment-service/config"
	"payment-service/internal/application"
	"payment-service/internal/domain"
	redislock "payment-service/internal/infra/lock"
	stripeinfra "payment-service/internal/infra/stripe"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// ServiceModule provides the payment application service.
var ServiceModule = fx.Options(
	fx.Provide(ProvideStripeConnectGateway, ProvideConnectAccountLock, ProvidePaymentService),
)

// ProvideConnectAccountLock wires the distributed onboarding lock.
func ProvideConnectAccountLock(client *redis.Client) (application.ConnectAccountLock, error) {
	return redislock.NewRedisConnectAccountLock(client)
}

// ProvidePaymentService constructs the payment application service.
func ProvidePaymentService(writeRepo domain.PaymentIntentRepository, webhooks domain.WebhookRepository, read domain.PaymentIntentReadRepository, accounts domain.ConnectAccountRepository, releases domain.PaymentReleaseRepository, broker application.EventBroker, stripe application.StripeConnectGateway, lock application.ConnectAccountLock, cfg *config.Config, lg logging.Logger) (application.Service, error) {
	return application.New(writeRepo, webhooks, read, accounts, releases, broker, cfg.Kafka.PaymentProjectionTopic, stripe, cfg.Stripe.FakeEnabled, lock, lg)
}

// ProvideStripeConnectGateway constructs the Stripe Connect adapter.
func ProvideStripeConnectGateway(cfg *config.Config, lg logging.Logger) (application.StripeConnectGateway, error) {
	if cfg.Stripe.FakeEnabled {
		return stripeinfra.NewFake(lg)
	}
	return stripeinfra.New(cfg.Stripe, lg)
}
