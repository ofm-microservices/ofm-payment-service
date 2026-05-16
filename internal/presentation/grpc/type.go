package grpc

import (
	"context"

	app "payment-service/internal/application"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
)

// Logger aliases the shared structured logger.
type Logger = logging.Logger

// Service is the onboarding and payment application boundary used by the gRPC
// adapter.
type Service = app.Service

// StartFreelancerOnboardingRequest aliases the transport onboarding request.
type StartFreelancerOnboardingRequest = app.StartFreelancerOnboardingCommand

// StartFreelancerOnboardingResponse aliases the transport onboarding response.
type StartFreelancerOnboardingResponse = app.StartFreelancerOnboardingResult

// Server exposes the payment onboarding gRPC lifecycle.
type Server interface {
	Start() error
	Shutdown(context.Context) error
}
