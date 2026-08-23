package kafka

import (
	"context"
	paymentflowv1 "github.com/ofm-microservices/ofm-common/proto/paymentflow/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"payment-service/config"
	app "payment-service/internal/application"
)

// PaymentIntentSubscriber consumes payment intent commands from Kafka.
type PaymentIntentSubscriber interface{ Subscribe(context.Context) error }
type paymentIntentSubscriber struct {
	broker  app.EventBroker
	service app.Service
	topic   string
}

// NewPaymentIntentSubscriber constructs the payment command subscriber.
func NewPaymentIntentSubscriber(broker app.EventBroker, service app.Service, cfg config.KafkaConfig) (PaymentIntentSubscriber, error) {
	return &paymentIntentSubscriber{broker: broker, service: service, topic: cfg.PaymentIntentTopic}, nil
}

func (s *paymentIntentSubscriber) Subscribe(ctx context.Context) error {
	return s.broker.Subscribe(ctx, s.topic, func(ctx context.Context, _ string, payload []byte) error {
		var cmd paymentflowv1.PaymentIntentCommand
		if err := protojson.Unmarshal(payload, &cmd); err != nil {
			return err
		}
		_, err := s.service.CreateIntent(ctx, app.CreateIntentCommand{IntentID: cmd.GetPaymentIntentId(), SagaID: cmd.GetSagaId(), OrderID: cmd.GetOrderId(), AmountCents: cmd.GetAmountCents(), Currency: cmd.GetCurrency(), Provider: cmd.GetProvider(), IdempotencyKey: cmd.GetIdempotencyKey(), RequestedAt: cmd.GetRequestedAt()})
		return err
	})
}
