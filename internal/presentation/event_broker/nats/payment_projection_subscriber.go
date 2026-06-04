package nats

import (
	"context"

	"payment-service/config"
	app "payment-service/internal/application"

	paymentflowv1 "github.com/ofm-microservices/ofm-common/proto/paymentflow/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// PaymentProjectionSubscriber consumes payment projection repair jobs.
type PaymentProjectionSubscriber interface {
	Subscribe(ctx context.Context) error
}

type paymentProjectionSubscriber struct {
	broker  app.EventBroker
	service app.Service
	cfg     config.NATSConfig
}

// NewPaymentProjectionSubscriber constructs the payment projection repair subscriber.
func NewPaymentProjectionSubscriber(broker app.EventBroker, service app.Service, cfg config.NATSConfig) (PaymentProjectionSubscriber, error) {
	return &paymentProjectionSubscriber{broker: broker, service: service, cfg: cfg}, nil
}

func (s *paymentProjectionSubscriber) Subscribe(ctx context.Context) error {
	return s.broker.Subscribe(ctx, s.cfg.PaymentProjectionSubject, func(ctx context.Context, subject string, payload []byte) error {
		_ = subject
		var msg paymentflowv1.PaymentProjectionRequest
		if err := protojson.Unmarshal(payload, &msg); err != nil {
			return err
		}
		return s.service.ProjectPaymentByOrderID(ctx, msg.GetOrderId())
	})
}
