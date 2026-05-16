package grpc

import "errors"

var (
	ErrNilService = errors.New("payment service is nil")
	ErrNilLogger  = errors.New("logger is nil")
	ErrInvalidRequestBody = errors.New("invalid request body")
)
