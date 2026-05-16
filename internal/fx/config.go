package appfx

import (
	"go.uber.org/fx"
	"payment-service/config"
)

// ConfigModule loads the payment-service configuration.
var ConfigModule = fx.Provide(config.Load)
