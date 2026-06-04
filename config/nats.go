package config

import "time"

// NATSConfig defines NATS streams and subjects used by payment-service.
type NATSConfig struct {
	URL      string `env:"NATS_URL,required"`
	User     string `env:"NATS_USER"`
	Password string `env:"NATS_PASSWORD"`

	PaymentCommandsStream      string `env:"NATS_STREAM_PAYMENT_COMMANDS" envDefault:"PAYMENT_COMMANDS"`
	PaymentEventsStream        string `env:"NATS_STREAM_PAYMENT_EVENTS" envDefault:"PAYMENT_EVENTS"`
	PaymentIntentSubject       string `env:"NATS_SUBJECT_PAYMENT_INTENT" envDefault:"payment.intent"`
	PaymentIntentResultSubject string `env:"NATS_SUBJECT_PAYMENT_INTENT_RESULT" envDefault:"payment.intent.result"`
	PaymentProjectionSubject   string `env:"NATS_SUBJECT_PAYMENT_PROJECTION" envDefault:"payment.projection"`
	PaymentWebhookSubject      string `env:"NATS_SUBJECT_PAYMENT_WEBHOOK" envDefault:"payment.webhook"`
	PaymentSucceededSubject    string `env:"NATS_SUBJECT_PAYMENT_SUCCEEDED" envDefault:"payment.order_payment_succeeded"`
	PaymentFailedSubject       string `env:"NATS_SUBJECT_PAYMENT_FAILED" envDefault:"payment.order_payment_failed"`
	PaymentCompensationSubject string `env:"NATS_SUBJECT_PAYMENT_COMPENSATION" envDefault:"payment.compensation"`

	CommandBatchSize int           `env:"NATS_COMMAND_BATCH_SIZE" envDefault:"32"`
	CommandMaxWait   time.Duration `env:"NATS_COMMAND_MAX_WAIT" envDefault:"10ms"`
}
