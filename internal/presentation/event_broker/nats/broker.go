package nats

import (
	"context"
	"fmt"
	"time"

	"payment-service/config"
	eb "payment-service/internal/presentation/event_broker"

	"github.com/nats-io/nats.go"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/natstrace"
)

type broker struct {
	conn *nats.Conn
	log  logging.Logger
}

// NewBroker connects to NATS and returns the runtime broker.
func NewBroker(cfg config.NATSConfig, log logging.Logger) (eb.EventBroker, error) {
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}
	opts := []nats.Option{nats.Timeout(5 * time.Second), nats.Name("payment-service"), nats.MaxReconnects(-1)}
	if cfg.User != "" || cfg.Password != "" {
		opts = append(opts, nats.UserInfo(cfg.User, cfg.Password))
	}
	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, WrapConnectToNATSError(err)
	}
	return &broker{conn: conn, log: log.With(logging.String("module", "nats-broker"))}, nil
}

func (b *broker) Publish(ctx context.Context, subject string, payload []byte) error {
	if err := b.conn.PublishMsg(natstrace.NewMessage(ctx, subject, payload)); err != nil {
		return WrapPublishToNATSError(subject, err)
	}
	return b.conn.Flush()
}

func (b *broker) Subscribe(_ context.Context, subject string, handler eb.MessageHandler) error {
	_, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {
		_ = handler(natstrace.ContextFromMessage(context.Background(), msg), msg.Subject, msg.Data)
	})
	if err != nil {
		return WrapSubscribeToNATSError(subject, err)
	}
	return b.conn.Flush()
}

func (b *broker) Close() {
	if b.conn != nil {
		b.conn.Close()
	}
}
