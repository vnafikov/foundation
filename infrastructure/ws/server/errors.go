package server

import "errors"

var (
	ErrConnectionClosed = errors.New("WebSocket connection is closed")
	ErrClosed           = errors.New("WebSocket server is closed")

	ErrMessageTooBig = errors.New("WebSocket message is too big")

	ErrEmptyChannelPrefix = errors.New("WebSocket channel prefix is empty")

	errReceiveMessageRejectionLimitReached = errors.New("WebSocket receive message rejection limit is reached")
	errReceiveMessageRateExceeded          = errors.New("WebSocket receive message rate is exceeded")

	errSendBufferFull = errors.New("WebSocket send buffer is full")
)
