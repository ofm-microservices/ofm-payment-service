package appfx

import (
	"context"

	"payment-service/config"
	app "payment-service/internal/application"
	eventbroker "payment-service/internal/presentation/event_broker"
	events "payment-service/internal/presentation/event_broker/nats"
	gatewaygrpc "payment-service/internal/presentation/grpc"
	webhookhttp "payment-service/internal/presentation/http"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"go.uber.org/fx"
)

// PresentationModule wires the HTTP webhook adapter and NATS background
// subscribers into the FX lifecycle.
var PresentationModule = fx.Options(
	fx.Provide(
		ProvidePaymentIntentSubscriber,
		ProvideOnboardingGRPCServer,
		ProvideWebhookServer,
	),
	fx.Invoke(
		InvokeSubscribePaymentIntent,
		InvokeRunOnboardingGRPCServer,
		InvokeRunWebhookServer,
	),
)

// ProvidePaymentIntentSubscriber constructs the NATS subscriber used to
// consume payment-intent commands.
func ProvidePaymentIntentSubscriber(broker eventbroker.EventBroker, service app.Service, cfg *config.Config, lg logging.Logger) (events.PaymentIntentSubscriber, error) {
	_ = lg
	return events.NewPaymentIntentSubscriber(broker, service, cfg.NATS)
}

// ProvideWebhookServer constructs the Stripe webhook HTTP server.
func ProvideWebhookServer(service app.Service, cfg *config.Config, lg logging.Logger) (webhookhttp.Server, error) {
	return webhookhttp.NewServer(service, cfg.HTTP, cfg.Stripe, lg)
}

// ProvideOnboardingGRPCServer constructs the onboarding gRPC server.
func ProvideOnboardingGRPCServer(service app.Service, cfg *config.Config, lg logging.Logger) (gatewaygrpc.Server, error) {
	return gatewaygrpc.NewServer(service, cfg.GRPC, lg)
}

// InvokeSubscribePaymentIntent starts the NATS subscriber with the FX
// lifecycle.
func InvokeSubscribePaymentIntent(lc fx.Lifecycle, subscriber events.PaymentIntentSubscriber, cfg *config.Config, lg logging.Logger) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			runCtx, runCancel := context.WithCancel(context.Background())
			cancel = runCancel
			if err := subscriber.Subscribe(runCtx); err != nil {
				lg.Error("subscribe to payment intent commands failed", logging.Err(err))
				cancel()
				return err
			}
			lg.Info("payment-service initialized", logging.String("env", cfg.App.Env))
			return nil
		},
		OnStop: func(context.Context) error {
			if cancel != nil {
				cancel()
			}
			return nil
		},
	})
}

// InvokeRunWebhookServer starts and gracefully stops the webhook HTTP server.
func InvokeRunWebhookServer(lc fx.Lifecycle, srv webhookhttp.Server) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.Start(); err != nil {
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error { return srv.Shutdown() },
	})
}

// InvokeRunOnboardingGRPCServer starts and gracefully stops the onboarding gRPC server.
func InvokeRunOnboardingGRPCServer(lc fx.Lifecycle, srv gatewaygrpc.Server) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.Start(); err != nil {
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error { return srv.Shutdown(ctx) },
	})
}
