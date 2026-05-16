package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"payment-service/config"
	app "payment-service/internal/application"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	paymentconnectv1 "github.com/ofm-microservices/ofm-common/proto/paymentconnect/v1"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	grpcpkg "google.golang.org/grpc"
)

type server struct {
	paymentconnectv1.UnimplementedPaymentOnboardingServiceServer
	svc      app.Service
	cfg      config.GRPCConfig
	log      logging.Logger
	srv      *grpcpkg.Server
	listener net.Listener
	mapr     *mapper
}

// NewServer constructs the payment onboarding gRPC server.
func NewServer(svc app.Service, cfg config.GRPCConfig, log logging.Logger) (Server, error) {
	if svc == nil {
		return nil, ErrNilService
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	grpcSrv := grpcpkg.NewServer(
		grpcpkg.StatsHandler(otelgrpc.NewServerHandler()),
		grpcpkg.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
	s := &server{
		svc:  svc,
		cfg:  cfg,
		log:  log.With(logging.String("module", "grpc-payment-onboarding-server")),
		srv:  grpcSrv,
		mapr: newMapper(),
	}
	paymentconnectv1.RegisterPaymentOnboardingServiceServer(grpcSrv, s)
	return s, nil
}

// Start begins serving gRPC traffic.
func (s *server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = lis
	s.log.Info("starting grpc server", logging.String("addr", addr))
	return s.srv.Serve(lis)
}

// Shutdown gracefully stops the gRPC server.
func (s *server) Shutdown(context.Context) error {
	if s.srv != nil {
		s.log.Info("shutting down grpc server")
		s.srv.GracefulStop()
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// StartFreelancerOnboarding provisions Stripe Connect onboarding for a user.
func (s *server) StartFreelancerOnboarding(ctx context.Context, req *paymentconnectv1.StartFreelancerOnboardingRequest) (*paymentconnectv1.StartFreelancerOnboardingResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.StartFreelancerOnboarding(ctx, s.mapr.ToStartFreelancerOnboardingRequest(req))
	if err != nil {
		log.Error("start freelancer onboarding failed",
			logging.Operation("grpc.payment.start_onboarding"),
			logging.DurationMS(time.Since(started)),
			logging.String("user_id", req.GetUserId()),
			logging.Err(err),
		)
		return nil, err
	}
	return s.mapr.ToStartFreelancerOnboardingResponse(res), nil
}
