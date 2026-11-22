package server

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/vnafikov/foundation/logger"
)

const messageRateExceededInterval = time.Second

type Connection struct {
	cancelCtx                    context.CancelFunc
	server                       *Server
	conn                         *websocket.Conn
	messageLimiterShard          *messageLimiterShard
	anonymousMessageLimiter      messageLimiter
	messageRejectionsStartedAt   time.Time
	messageRejectionCount        int
	messageRateExceededHandledAt time.Time
	recipient                    string
	subscriptions                map[string]struct{}
	sendMessageChan              chan []byte
	canceledChan                 <-chan struct{}
	doneChan                     chan struct{}
	mu                           sync.Mutex
}

func (c *Connection) run(ctx context.Context) {
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go c.read(ctx, wg)

	defer func() {
		c.cancelCtx()
		wg.Wait()
	}()

	ticker := time.NewTicker(c.server.configuration.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-c.sendMessageChan:
			if err := c.write(ctx, payload, c.server.configuration.WriteTimeout); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.ping(ctx, c.server.configuration.PingTimeout); err != nil {
				return
			}
		}
	}
}

/*
read читает сообщения клиента и последовательно вызывает обработчик.

Без обработчика закрывает соединение при первом текстовом или бинарном сообщении, не читая его тело.
При ошибке чтения или обработчика закрывает соединение.
При превышении лимита отклоняет сообщение
и при наличии [Configuration.HandleMessageRateExceeded] вызывает его не чаще одного раза в секунду для соединения.
При достижении [Configuration.ReceiveMessageRejectionLimit] за [Configuration.ReceiveMessageRejectionInterval] закрывает соединение.
Успешные сообщения не сбрасывают счётчик отказов.
*/
func (c *Connection) read(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		c.cancelCtx()
		wg.Done()
	}()

	if c.server.configuration.HandleMessage == nil {
		_, _, err := c.conn.Reader(ctx)
		if err != nil && logger.ShouldLogDebug() {
			logger.DebugError(ctx, err, "websocket.(*Conn).Reader", slog.String("recipient", c.recipient))
		}
		return
	}

	for {
		_, payload, err := c.conn.Read(ctx)
		if err != nil {
			if logger.ShouldLogDebug() {
				logger.DebugError(ctx, err, "websocket.(*Conn).Read", slog.String("recipient", c.recipient))
			}
			return
		}

		if !c.allowReceiveMessage() {
			if err = c.handleMessageRateExceeded(ctx); err != nil {
				return
			}
			continue
		}

		if err = c.server.configuration.HandleMessage(ctx, c, payload); err != nil {
			return
		}
	}
}

func (c *Connection) allowReceiveMessage() bool {
	if c.server.configuration.ReceiveMessageRate <= 0 || c.server.configuration.ReceiveMessageBurst <= 0 {
		return true
	}

	interval := time.Second / time.Duration(c.server.configuration.ReceiveMessageRate)
	burstInterval := interval * time.Duration(c.server.configuration.ReceiveMessageBurst)
	if c.messageLimiterShard != nil {
		return c.messageLimiterShard.allow(c.recipient, interval, burstInterval)
	}
	return c.anonymousMessageLimiter.allow(time.Now(), interval, burstInterval)
}

func (c *Connection) handleMessageRateExceeded(ctx context.Context) error {
	now := time.Now()
	if c.server.configuration.ReceiveMessageRejectionLimit > 0 && c.server.configuration.ReceiveMessageRejectionInterval > 0 {
		if c.messageRejectionCount == 0 || now.Sub(c.messageRejectionsStartedAt) >= c.server.configuration.ReceiveMessageRejectionInterval {
			c.messageRejectionsStartedAt = now
			c.messageRejectionCount = 0
		}
		c.messageRejectionCount++
		if c.messageRejectionCount >= c.server.configuration.ReceiveMessageRejectionLimit {
			c.cancelCtx()
			if logger.ShouldLogDebug() {
				logger.DebugError(
					ctx,
					errReceiveMessageRejectionLimitReached,
					"server.(*Connection).handleMessageRateExceeded",
					slog.String("recipient", c.recipient),
				)
			}
			return ErrConnectionClosed
		}
	}
	if now.Sub(c.messageRateExceededHandledAt) < messageRateExceededInterval {
		return nil
	}

	c.messageRateExceededHandledAt = now
	if logger.ShouldLogDebug() {
		logger.DebugError(
			ctx,
			errReceiveMessageRateExceeded,
			"server.(*Connection).handleMessageRateExceeded",
			slog.String("recipient", c.recipient),
		)
	}
	if c.server.configuration.HandleMessageRateExceeded != nil {
		return c.server.configuration.HandleMessageRateExceeded(ctx, c)
	}
	return nil
}

func (c *Connection) write(ctx context.Context, payload []byte, timeout time.Duration) error {
	writeCtx, writeCancelCtx := context.WithTimeout(ctx, timeout)
	defer writeCancelCtx()

	err := c.conn.Write(writeCtx, websocket.MessageText, payload)
	if err != nil && logger.ShouldLogDebug() {
		logger.DebugError(writeCtx, err, "websocket.(*Conn).Write", slog.String("recipient", c.recipient))
	}
	return err
}

func (c *Connection) ping(ctx context.Context, timeout time.Duration) error {
	pingCtx, pingCancelCtx := context.WithTimeout(ctx, timeout)
	defer pingCancelCtx()

	err := c.conn.Ping(pingCtx)
	if err != nil && logger.ShouldLogDebug() {
		logger.DebugError(pingCtx, err, "websocket.(*Conn).Ping", slog.String("recipient", c.recipient))
	}
	return err
}

/*
Subscribe подписывает соединение на канал.

Повторная подписка ничего не меняет.

IMPORTANT: Перед вызовом необходимо проверить права на подписку.
*/
func (c *Connection) Subscribe(channel string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.canceledChan:
		return ErrConnectionClosed
	default:
	}

	shard := c.server.subscriptionShard(channel)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if shard.isClosed {
		return ErrClosed
	}

	if c.subscriptions == nil {
		c.subscriptions = make(map[string]struct{})
	}
	shard.addLocked(channel, c)
	c.subscriptions[channel] = struct{}{}
	return nil
}

/*
Unsubscribe отписывает соединение от канала и возвращает, снята ли подписка.

Уже помещённые в буфер сообщения не отзывает.
*/
func (c *Connection) Unsubscribe(channel string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.subscriptions[channel]; !ok {
		return false
	}

	c.unsubscribeLocked(channel)
	return true
}

func (c *Connection) unsubscribeLocked(channel string) {
	c.unregisterSubscription(channel)
	delete(c.subscriptions, channel)
}

func (c *Connection) unregisterSubscription(channel string) {
	shard := c.server.subscriptionShard(channel)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.deleteLocked(channel, c)
}

func (c *Connection) unsubscribeByPrefix(channelPrefix string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var channels []string
	for channel := range c.subscriptions {
		if !strings.HasPrefix(channel, channelPrefix) {
			continue
		}

		c.unsubscribeLocked(channel)
		channels = append(channels, channel)
	}
	return channels
}

/*
Send передаёт текстовое сообщение в буфер соединения.

При переполнении буфера закрывает соединение.

IMPORTANT: После вызова нельзя изменять содержимое payload.
*/
func (c *Connection) Send(ctx context.Context, payload []byte) error {
	if err := c.server.validateMessage(ctx, payload); err != nil {
		return err
	}
	return c.enqueue(ctx, payload)
}

func (c *Connection) enqueue(ctx context.Context, payload []byte) error {
	select {
	case <-c.canceledChan:
		return ErrConnectionClosed
	default:
	}

	select {
	case c.sendMessageChan <- payload:
		return nil
	default:
		c.cancelCtx()
		if logger.ShouldLogDebug() {
			logger.DebugError(ctx, errSendBufferFull, "server.(*Connection).enqueue", slog.String("recipient", c.recipient))
		}
		return ErrConnectionClosed
	}
}

func (c *Connection) unsubscribeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for channel := range c.subscriptions {
		c.unregisterSubscription(channel)
	}
	clear(c.subscriptions)
}

// Recipient возвращает получателя, указанного при открытии соединения.
func (c *Connection) Recipient() string {
	return c.recipient
}

func newConnection(
	cancelCtx context.CancelFunc,
	server *Server,
	conn *websocket.Conn,
	recipient string,
	canceledChan <-chan struct{},
) *Connection {
	return &Connection{
		cancelCtx:       cancelCtx,
		server:          server,
		conn:            conn,
		recipient:       recipient,
		sendMessageChan: make(chan []byte, server.configuration.SendBufferCapacity),
		canceledChan:    canceledChan,
		doneChan:        make(chan struct{}),
	}
}
