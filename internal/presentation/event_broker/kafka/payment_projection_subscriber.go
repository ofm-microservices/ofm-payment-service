package kafka

import (
	"context"
	paymentflowv1 "github.com/ofm-microservices/ofm-common/proto/paymentflow/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"payment-service/config"
	app "payment-service/internal/application"
)

// PaymentProjectionSubscriber consumes payment projection repair jobs.
type PaymentProjectionSubscriber interface{ Subscribe(context.Context) error }
type paymentProjectionSubscriber struct {
	broker  app.EventBroker
	service app.Service
	topic   string
}

// NewPaymentProjectionSubscriber constructs the payment projection subscriber.
func NewPaymentProjectionSubscriber(broker app.EventBroker, service app.Service, cfg config.KafkaConfig) (PaymentProjectionSubscriber, error) {
	return &paymentProjectionSubscriber{broker: broker, service: service, topic: cfg.PaymentProjectionTopic}, nil
}

func (s *paymentProjectionSubscriber) Subscribe(ctx context.Context) error {
	return s.broker.Subscribe(ctx, s.topic, func(ctx context.Context, _ string, payload []byte) error {
		var msg paymentflowv1.PaymentProjectionRequest
		if err := protojson.Unmarshal(payload, &msg); err != nil {
			return err
		}
		return s.service.ProjectPaymentByOrderID(ctx, msg.GetOrderId())
	})
}
