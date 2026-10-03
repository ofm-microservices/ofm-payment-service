package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config groups payment-service runtime settings.
type Config struct {
	App     AppConfig
	DB      DBConfig
	Redis   RedisConfig
	Kafka   KafkaConfig
	HTTP    HTTPConfig
	GRPC    GRPCConfig
	Metrics MetricsConfig
	Tracing TracingConfig
	Stripe  StripeConfig
}

// Load reads environment variables into Config.
func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, WrapParseEnvConfigError(err)
	}
	return cfg, nil
}
