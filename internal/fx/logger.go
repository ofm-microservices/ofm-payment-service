package appfx

import (
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"go.uber.org/fx"
	"payment-service/config"
)

// LoggerModule provides the structured logger.
var LoggerModule = fx.Provide(ProvideLogger)

func ProvideLogger(cfg *config.Config) (logging.Logger, error) {
	return logging.New(cfg.App.Name, cfg.App.Env, cfg.App.LogLevel)
}
