package domain

import "errors"

var (
	ErrInvalidID              = errors.New("invalid id")
	ErrIntentNotFound         = errors.New("payment intent not found")
	ErrPaymentReleaseNotFound = errors.New("payment release not found")
	ErrWebhookNotFound        = errors.New("payment webhook not found")
	ErrDedupNotFound          = errors.New("payment dedup record not found")
	ErrConnectAccountNotFound = errors.New("payment connect account not found")
)
