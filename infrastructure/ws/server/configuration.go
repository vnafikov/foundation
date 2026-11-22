package server

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

type Configuration struct {
	HandleMessage             func(ctx context.Context, connection *Connection, payload []byte) error
	HandleMessageRateExceeded func(ctx context.Context, connection *Connection) error
	OriginPatterns            []string
	CompressionMode           websocket.CompressionMode
	SendBufferCapacity        int
	MaxMessageSize            int64
	ConnectionTTL             time.Duration
	WriteTimeout              time.Duration
	PingInterval              time.Duration
	PingTimeout               time.Duration
	// ReceiveMessageRate задаёт число входящих сообщений в секунду на получателя, для анонимных на соединение.
	// При нулевом ReceiveMessageRate или ReceiveMessageBurst ограничение отключено.
	ReceiveMessageRate int
	// ReceiveMessageBurst задаёт максимальный запас сообщений для отправки подряд.
	ReceiveMessageBurst int
	// ReceiveMessageRejectionLimit задаёт число отказов за ReceiveMessageRejectionInterval для отключения соединения.
	// При нулевом ReceiveMessageRejectionLimit или ReceiveMessageRejectionInterval отключение не применяется.
	ReceiveMessageRejectionLimit int
	// ReceiveMessageRejectionInterval задаёт окно подсчёта отказов, начиная с первого отказа.
	ReceiveMessageRejectionInterval time.Duration
}
