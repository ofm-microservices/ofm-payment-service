package grpc

import (
	"context"

	app "payment-service/internal/application"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	paymentcheckoutv1 "github.com/ofm-microservices/ofm-common/proto/paymentcheckout/v1"
	paymentconnectv1 "github.com/ofm-microservices/ofm-common/proto/paymentconnect/v1"
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

// GetConnectStatusRequest aliases the transport connect status request.
type GetConnectStatusRequest = paymentconnectv1.GetConnectStatusRequest

// GetConnectStatusResponse aliases the transport connect status response.
type GetConnectStatusResponse = paymentconnectv1.GetConnectStatusResponse

// GetPaymentByOrderRequest aliases the transport payment-by-order request.
type GetPaymentByOrderRequest = paymentcheckoutv1.GetPaymentByOrderIdRequest

// GetPaymentByOrderResponse aliases the transport payment-by-order response.
type GetPaymentByOrderResponse = paymentcheckoutv1.GetPaymentByOrderIdResponse

// Server exposes the payment onboarding gRPC lifecycle.
type Server interface {
	Start() error
	Shutdown(context.Context) error
}
