package config

// HTTPConfig defines the webhook listener settings.
type HTTPConfig struct {
	Host string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"HTTP_PORT" envDefault:"8081"`
}
