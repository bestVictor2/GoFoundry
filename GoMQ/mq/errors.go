package mq

import "errors"

var (
	ErrURLRequired     = errors.New("mq url is required")
	ErrQueueRequired   = errors.New("mq queue is required")
	ErrClientClosed    = errors.New("mq client is closed")
	ErrHandlerRequired = errors.New("mq handler is required")
)
