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
	PaymentWebhookSubject      string `env:"NATS_SUBJECT_PAYMENT_WEBHOOK" envDefault:"payment.webhook"`
	PaymentStatusSubject       string `env:"NATS_SUBJECT_PAYMENT_STATUS" envDefault:"payment.status"`
	PaymentCompletedSubject    string `env:"NATS_SUBJECT_PAYMENT_COMPLETED" envDefault:"payment.completed"`
	PaymentFailedSubject       string `env:"NATS_SUBJECT_PAYMENT_FAILED" envDefault:"payment.failed"`
	PaymentCompensationSubject string `env:"NATS_SUBJECT_PAYMENT_COMPENSATION" envDefault:"payment.compensation"`

	CommandBatchSize int           `env:"NATS_COMMAND_BATCH_SIZE" envDefault:"32"`
	CommandMaxWait   time.Duration `env:"NATS_COMMAND_MAX_WAIT" envDefault:"10ms"`
}
