package stripe

import "errors"

var (
	ErrNilLogger            = errors.New("logger is nil")
	ErrCreateAccount        = errors.New("create stripe connect account failed")
	ErrCreateOnboardingLink = errors.New("create stripe onboarding link failed")
	ErrCreateRefund         = errors.New("create stripe refund failed")
)
