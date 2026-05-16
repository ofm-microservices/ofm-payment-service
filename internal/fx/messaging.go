package appfx

import (
	"context"

	"payment-service/config"
	eventbroker "payment-service/internal/presentation/event_broker"
	broker "payment-service/internal/presentation/event_broker/nats"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"go.uber.org/fx"
	natsbootstrap "payment-service/pkg/messaging/nats"
)

// MessagingModule wires NATS bootstrap and broker runtime into payment-service.
var MessagingModule = fx.Options(
	fx.Invoke(InvokeEnsureStream),
	fx.Provide(ProvideEventBroker),
)

// InvokeEnsureStream ensures the JetStream streams payment-service depends on.
func InvokeEnsureStream(cfg *config.Config, lg logging.Logger) error {
	if err := natsbootstrap.EnsureStream(cfg.NATS, lg); err != nil {
		lg.Error("bootstrap jetstream resources failed", logging.Err(err))
		return err
	}
	return nil
}

// ProvideEventBroker constructs the concrete NATS event broker.
func ProvideEventBroker(lc fx.Lifecycle, cfg *config.Config, lg logging.Logger) (eventbroker.EventBroker, error) {
	eventBroker, err := broker.NewBroker(cfg.NATS, lg)
	if err != nil {
		lg.Error("connect nats failed", logging.Err(err))
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { eventBroker.Close(); return nil }})
	return eventBroker, nil
}
