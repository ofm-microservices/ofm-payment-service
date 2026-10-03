package config

// KafkaConfig defines Kafka topics and consumer settings owned by payment-service.
type KafkaConfig struct {
	Brokers                []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"127.0.0.1:9092"`
	GroupID                string   `env:"KAFKA_PAYMENT_GROUP_ID" envDefault:"payment-service"`
	PaymentIntentTopic     string   `env:"KAFKA_PAYMENT_INTENT_TOPIC" envDefault:"payment.intent"`
	PaymentProjectionTopic string   `env:"KAFKA_PAYMENT_PROJECTION_TOPIC" envDefault:"payment.projection"`
	PaymentIndexedTopic    string   `env:"KAFKA_PAYMENT_INDEXED_TOPIC" envDefault:"payment.intent.result"`
	RecoveryTopic          string   `env:"KAFKA_PAYMENT_RECOVERY_TOPIC" envDefault:"migration.recovery.commands.payment"`
	RecoveryGroup          string   `env:"KAFKA_PAYMENT_RECOVERY_GROUP" envDefault:"payment-service-recovery"`
	RecoveryCompletedTopic string   `env:"KAFKA_PAYMENT_RECOVERY_COMPLETED_TOPIC" envDefault:"migration.recovery.completed"`
}
