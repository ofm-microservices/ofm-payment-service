package http

import (
	"fmt"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"payment-service/config"
	app "payment-service/internal/application"

	"github.com/gofiber/fiber/v2"
)

// Server owns the webhook HTTP listener.
type Server interface {
	Start() error
	Shutdown() error
}

type server struct {
	app    app.Service
	cfg    config.HTTPConfig
	stripe config.StripeConfig
	log    logging.Logger
	fapp   *fiber.App
}

// NewServer constructs the Stripe webhook server.
func NewServer(svc app.Service, cfg config.HTTPConfig, stripeCfg config.StripeConfig, log logging.Logger) (Server, error) {
	if svc == nil {
		return nil, fmt.Errorf("payment service is nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}
	fapp := fiber.New()
	s := &server{app: svc, cfg: cfg, stripe: stripeCfg, log: log.With(logging.String("module", "http-webhook")), fapp: fapp}
	fapp.Post("/v1/payments/webhook", s.handleWebhook)
	fapp.Post("/v1/freelancer/onboarding/webhook", s.handleConnectWebhook)
	return s, nil
}

func (s *server) Start() error    { return s.fapp.Listen(fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)) }
func (s *server) Shutdown() error { return s.fapp.Shutdown() }
