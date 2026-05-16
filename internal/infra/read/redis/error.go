package redis

import "errors"

var (
	ErrNilRedisClient   = errors.New("redis client is nil")
	ErrNilLogger        = errors.New("logger is nil")
	ErrNilPaymentIntent = errors.New("payment intent is nil")
)
