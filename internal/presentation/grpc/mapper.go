package grpc

import (
	app "payment-service/internal/application"

	paymentconnectv1 "github.com/ofm-microservices/ofm-common/proto/paymentconnect/v1"
)

type mapper struct{}

func newMapper() *mapper { return &mapper{} }

func (m *mapper) ToStartFreelancerOnboardingRequest(req *paymentconnectv1.StartFreelancerOnboardingRequest) app.StartFreelancerOnboardingCommand {
	return app.StartFreelancerOnboardingCommand{
		UserID:         req.GetUserId(),
		Country:        req.GetCountry(),
		ReturnURL:      req.GetReturnUrl(),
		RefreshURL:     req.GetRefreshUrl(),
		IdempotencyKey: req.GetIdempotencyKey(),
		RequestedAt:    req.GetRequestedAt(),
	}
}

func (m *mapper) ToStartFreelancerOnboardingResponse(res *app.StartFreelancerOnboardingResult) *paymentconnectv1.StartFreelancerOnboardingResponse {
	return &paymentconnectv1.StartFreelancerOnboardingResponse{
		UserId:           res.UserID,
		StripeAccountId:  res.StripeAccountID,
		OnboardingUrl:    res.OnboardingURL,
		Status:           res.Status,
		DetailsSubmitted: res.DetailsSubmitted,
		ChargesEnabled:   res.ChargesEnabled,
		PayoutsEnabled:   res.PayoutsEnabled,
		DisabledReason:   res.DisabledReason,
		OccurredAt:       res.OccurredAt,
	}
}

func (m *mapper) ToGetConnectStatusResponse(res *app.GetConnectStatusResult) *paymentconnectv1.GetConnectStatusResponse {
	return &paymentconnectv1.GetConnectStatusResponse{
		UserId:          res.UserID,
		Status:          res.Status,
		StripeAccountId: res.StripeAccountID,
		DisabledReason:  res.DisabledReason,
		OccurredAt:      res.OccurredAt,
	}
}
