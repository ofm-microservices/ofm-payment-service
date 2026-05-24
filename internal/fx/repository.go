package appfx

import (
	"payment-service/internal/domain"
	readrepo "payment-service/internal/infra/read/redis"
	writerepo "payment-service/internal/infra/write/yugabyte"

	"github.com/jmoiron/sqlx"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// RepoModule wires write- and read-model repositories into the FX graph.
var RepoModule = fx.Options(
	fx.Provide(writerepo.NewDBErrorTranslator, ProvideWriteRepo, ProvideWebhookRepo, ProvideConnectAccountRepo, ProvidePaymentReleaseRepo, ProvideReadRepo),
)

// ProvideWriteRepo constructs the Yugabyte-backed payment repository.
func ProvideWriteRepo(dbx *sqlx.DB, translator writerepo.DBErrorTranslator, lg logging.Logger) (domain.PaymentIntentRepository, error) {
	return writerepo.New(dbx, translator, lg)
}

// ProvideWebhookRepo constructs the Yugabyte-backed payment webhook repository.
func ProvideWebhookRepo(dbx *sqlx.DB, translator writerepo.DBErrorTranslator, lg logging.Logger) (domain.WebhookRepository, error) {
	return writerepo.NewWebhookRepo(dbx, translator, lg)
}

// ProvideConnectAccountRepo constructs the Yugabyte-backed connect-account repository.
func ProvideConnectAccountRepo(dbx *sqlx.DB, translator writerepo.DBErrorTranslator, lg logging.Logger) (domain.ConnectAccountRepository, error) {
	return writerepo.NewConnectAccountRepo(dbx, translator, lg)
}

// ProvidePaymentReleaseRepo constructs the Yugabyte-backed payout release repository.
func ProvidePaymentReleaseRepo(dbx *sqlx.DB, translator writerepo.DBErrorTranslator, lg logging.Logger) (domain.PaymentReleaseRepository, error) {
	return writerepo.NewPaymentReleaseRepo(dbx, translator, lg)
}

// ProvideReadRepo constructs the Redis-backed payment read repository.
func ProvideReadRepo(rdb *redis.Client, lg logging.Logger) (domain.PaymentIntentReadRepository, error) {
	return readrepo.New(rdb, lg)
}
