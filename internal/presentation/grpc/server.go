package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"payment-service/config"
	app "payment-service/internal/application"

	"github.com/google/uuid"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	paymentcheckoutv1 "github.com/ofm-microservices/ofm-common/proto/paymentcheckout/v1"
	paymentconnectv1 "github.com/ofm-microservices/ofm-common/proto/paymentconnect/v1"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	grpcpkg "google.golang.org/grpc"
)

type server struct {
	paymentconnectv1.UnimplementedPaymentOnboardingServiceServer
	paymentcheckoutv1.UnimplementedPaymentCheckoutServiceServer
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
	paymentcheckoutv1.RegisterPaymentCheckoutServiceServer(grpcSrv, s)
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

// GetConnectStatus returns the current Connect onboarding state for one user.
func (s *server) GetConnectStatus(ctx context.Context, req *paymentconnectv1.GetConnectStatusRequest) (*paymentconnectv1.GetConnectStatusResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.GetConnectStatus(ctx, req.GetUserId())
	if err != nil {
		log.Error("get connect status failed",
			logging.Operation("grpc.payment.get_connect_status"),
			logging.DurationMS(time.Since(started)),
			logging.String("user_id", req.GetUserId()),
			logging.Err(err),
		)
		return nil, err
	}
	return s.mapr.ToGetConnectStatusResponse(res), nil
}

// CreateCheckoutSession creates a platform-held checkout session for an order.
func (s *server) CreateCheckoutSession(ctx context.Context, req *paymentcheckoutv1.CreateCheckoutSessionRequest) (*paymentcheckoutv1.CreateCheckoutSessionResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.CreateIntent(ctx, app.CreateIntentCommand{
		IntentID:       uuid.Must(uuid.NewV7()).String(),
		SagaID:         req.GetSagaId(),
		OrderID:        req.GetOrderId(),
		AmountCents:    req.GetAmountCents(),
		Currency:       req.GetCurrency(),
		Provider:       "stripe",
		IdempotencyKey: req.GetIdempotencyKey(),
		RequestedAt:    req.GetRequestedAt(),
	})
	if err != nil {
		log.Error("create checkout session failed",
			logging.Operation("grpc.payment.create_checkout_session"),
			logging.DurationMS(time.Since(started)),
			logging.String("order_id", req.GetOrderId()),
			logging.Err(err),
		)
		return nil, err
	}
	return &paymentcheckoutv1.CreateCheckoutSessionResponse{
		OrderId:               req.GetOrderId(),
		PaymentId:             res.IntentID,
		CheckoutUrl:           res.CheckoutURL,
		StripePaymentIntentId: res.ProviderIntentID,
		Status:                res.Status,
		OccurredAt:            res.OccurredAt,
	}, nil
}

// ReleaseFunds releases captured funds to the seller.
func (s *server) ReleaseFunds(ctx context.Context, req *paymentcheckoutv1.ReleaseFundsRequest) (*paymentcheckoutv1.ReleaseFundsResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.ReleaseFunds(ctx, app.ReleaseFundsCommand{
		OrderID:        req.GetOrderId(),
		PaymentID:      req.GetPaymentId(),
		SellerUserID:   req.GetSellerUserId(),
		AmountCents:    req.GetAmountCents(),
		Currency:       req.GetCurrency(),
		IdempotencyKey: req.GetIdempotencyKey(),
		RequestedAt:    req.GetRequestedAt(),
	})
	if err != nil {
		log.Error("release funds failed",
			logging.Operation("grpc.payment.release_funds"),
			logging.DurationMS(time.Since(started)),
			logging.String("order_id", req.GetOrderId()),
			logging.Err(err),
		)
		return nil, err
	}
	return &paymentcheckoutv1.ReleaseFundsResponse{
		OrderId:          res.OrderID,
		PaymentReleaseId: res.PaymentReleaseID,
		StripeTransferId: res.StripeTransferID,
		Status:           res.Status,
		OccurredAt:       res.OccurredAt,
	}, nil
}

// GetReleaseByOrderId returns the persisted payout release for recovery.
func (s *server) GetReleaseByOrderId(ctx context.Context, req *paymentcheckoutv1.GetReleaseByOrderIdRequest) (*paymentcheckoutv1.GetReleaseByOrderIdResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.GetReleaseByOrderID(ctx, req.GetOrderId())
	if err != nil {
		log.Error("get release by order id failed",
			logging.Operation("grpc.payment.get_release_by_order_id"),
			logging.DurationMS(time.Since(started)),
			logging.String("order_id", req.GetOrderId()),
			logging.Err(err),
		)
		return nil, err
	}
	return &paymentcheckoutv1.GetReleaseByOrderIdResponse{
		OrderId:          res.OrderID,
		PaymentReleaseId: res.PaymentReleaseID,
		PaymentIntentId:  res.PaymentID,
		SellerUserId:     res.SellerUserID,
		AmountCents:      res.AmountCents,
		Currency:         res.Currency,
		IdempotencyKey:   res.IdempotencyKey,
		StripeTransferId: res.StripeTransferID,
		Status:           res.Status,
		FailureReason:    res.FailureReason,
		OccurredAt:       res.OccurredAt,
	}, nil
}

// GetPaymentByOrderId returns the public payment snapshot for an order.
func (s *server) GetPaymentByOrderId(ctx context.Context, req *paymentcheckoutv1.GetPaymentByOrderIdRequest) (*paymentcheckoutv1.GetPaymentByOrderIdResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.GetPaymentByOrderID(ctx, req.GetOrderId())
	if err != nil {
		log.Error("get payment by order id failed",
			logging.Operation("grpc.payment.get_payment_by_order_id"),
			logging.DurationMS(time.Since(started)),
			logging.String("order_id", req.GetOrderId()),
			logging.Err(err),
		)
		return nil, err
	}
	return s.mapr.ToGetPaymentByOrderResponse(res), nil
}
