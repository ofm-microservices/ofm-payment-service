package application

import "errors"

var (
	ErrNilIntentRepository         = errors.New("payment intent repository is nil")
	ErrNilWebhookRepository        = errors.New("webhook repository is nil")
	ErrNilReadRepository           = errors.New("payment read repository is nil")
	ErrNilEventBroker              = errors.New("event broker is nil")
	ErrNilProjectionSubject        = errors.New("payment projection subject is nil")
	ErrNilConnectAccountRepository = errors.New("connect account repository is nil")
	ErrNilPaymentReleaseRepository = errors.New("payment release repository is nil")
	ErrNilStripeConnectGateway     = errors.New("stripe connect gateway is nil")
	ErrNilLogger                   = errors.New("logger is nil")
	ErrPublishEvent                = errors.New("publish event failed")
)
