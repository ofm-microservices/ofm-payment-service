package nats

import (
	"context"

	"payment-service/config"
	app "payment-service/internal/application"

	paymentflowv1 "github.com/ofm-microservices/ofm-common/proto/paymentflow/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// PaymentIntentSubscriber consumes payment intent commands from NATS.
type PaymentIntentSubscriber interface {
	Subscribe(ctx context.Context) error
}

type paymentIntentSubscriber struct {
	broker  app.EventBroker
	service app.Service
	cfg     config.NATSConfig
}

// NewPaymentIntentSubscriber constructs the payment command subscriber.
func NewPaymentIntentSubscriber(broker app.EventBroker, service app.Service, cfg config.NATSConfig) (PaymentIntentSubscriber, error) {
	return &paymentIntentSubscriber{broker: broker, service: service, cfg: cfg}, nil
}

func (s *paymentIntentSubscriber) Subscribe(ctx context.Context) error {
	return s.broker.Subscribe(ctx, s.cfg.PaymentIntentSubject, func(ctx context.Context, subject string, payload []byte) error {
		_ = subject
		var cmd paymentflowv1.PaymentIntentCommand
		if err := protojson.Unmarshal(payload, &cmd); err != nil {
			return err
		}
		_, err := s.service.CreateIntent(ctx, app.CreateIntentCommand{
			IntentID:       cmd.GetPaymentIntentId(),
			SagaID:         cmd.GetSagaId(),
			OrderID:        cmd.GetOrderId(),
			AmountCents:    cmd.GetAmountCents(),
			Currency:       cmd.GetCurrency(),
			Provider:       cmd.GetProvider(),
			IdempotencyKey: cmd.GetIdempotencyKey(),
			RequestedAt:    cmd.GetRequestedAt(),
		})
		return err
	})
}
