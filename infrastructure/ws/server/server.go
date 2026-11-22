package server

import (
	"context"
	"errors"
	"hash/maphash"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/coder/websocket"

	"github.com/vnafikov/foundation/infrastructure/http/server/response/servererrors"
	"github.com/vnafikov/foundation/logger"
)

const registryShardCount = 32

type Server struct {
	configuration            Configuration
	connectionShards         [registryShardCount]registryShard
	anonymousConnectionShard registryShard
	subscriptionShards       [registryShardCount]registryShard
	messageLimiterShards     [registryShardCount]messageLimiterShard
	seed                     maphash.Seed
	wg                       sync.WaitGroup
}

/*
Accept принимает WebSocket-соединение получателя и обслуживает его до закрытия.

Обрабатывает служебные кадры и передаёт прикладные сообщения клиента в [Configuration.HandleMessage].
Закрывает соединение при отмене контекста или истечении [Configuration.ConnectionTTL].
*/
func (s *Server) Accept(ctx context.Context, w http.ResponseWriter, r *http.Request, recipient string) {
	shard := s.connectionShard(recipient)
	s.acceptConnection(ctx, w, r, recipient, shard)
}

func (s *Server) connectionShard(recipient string) *registryShard {
	return &s.connectionShards[s.shardIndex(recipient)]
}

func (s *Server) shardIndex(key string) uint64 {
	return maphash.String(s.seed, key) % registryShardCount
}

func (s *Server) acceptConnection(ctx context.Context, w http.ResponseWriter, r *http.Request, recipient string, shard *registryShard) {
	shard.mu.Lock()
	isClosed := shard.isClosed
	shard.mu.Unlock()
	if isClosed || ctx.Err() != nil {
		servererrors.ServiceUnavailable(w)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  s.configuration.OriginPatterns,
		CompressionMode: s.configuration.CompressionMode,
	})
	if err != nil {
		if logger.ShouldLogDebug() {
			logger.DebugError(ctx, err, "websocket.Accept", slog.String("recipient", recipient))
		}
		return
	}

	conn.SetReadLimit(s.configuration.MaxMessageSize)
	connectionCtx, connectionCancelCtx := context.WithTimeout(ctx, s.configuration.ConnectionTTL)
	c := newConnection(connectionCancelCtx, s, conn, recipient, connectionCtx.Done())
	if shard != &s.anonymousConnectionShard {
		c.messageLimiterShard = s.messageLimiterShard(recipient)
	}
	isRegistered := s.register(c, shard)
	defer func() {
		connectionCancelCtx()
		err := conn.CloseNow()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Error(connectionCtx, err, "websocket.(*Conn).CloseNow", slog.String("recipient", recipient))
		}
		if !isRegistered {
			return
		}

		s.unregister(c, shard)
		close(c.doneChan)
		s.wg.Done()
	}()

	if !isRegistered {
		return
	}

	c.run(connectionCtx)
}

func (s *Server) messageLimiterShard(recipient string) *messageLimiterShard {
	return &s.messageLimiterShards[s.shardIndex(recipient)]
}

func (s *Server) register(c *Connection, shard *registryShard) bool {
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if shard.isClosed {
		return false
	}

	shard.addLocked(c.recipient, c)
	s.wg.Add(1)
	return true
}

func (*Server) unregister(c *Connection, shard *registryShard) {
	c.unsubscribeAll()
	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.deleteLocked(c.recipient, c)
}

/*
AcceptAnonymous принимает WebSocket-соединение без получателя и обслуживает его до закрытия.

Обрабатывает служебные кадры и передаёт прикладные сообщения клиента в [Configuration.HandleMessage].
Закрывает соединение при отмене контекста или истечении [Configuration.ConnectionTTL].
*/
func (s *Server) AcceptAnonymous(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	s.acceptConnection(ctx, w, r, "", &s.anonymousConnectionShard)
}

/*
Send передаёт текстовое сообщение в буферы текущих соединений получателя.

При отсутствии соединений пропускает отправку без ошибки.
При переполнении буфера отключает соединение.

IMPORTANT: После вызова нельзя изменять содержимое payload.
*/
func (s *Server) Send(ctx context.Context, recipient string, payload []byte) error {
	if err := s.validateMessage(ctx, payload); err != nil {
		return err
	}
	return s.connectionShard(recipient).send(ctx, recipient, payload)
}

func (s *Server) validateMessage(ctx context.Context, payload []byte) error {
	size := len(payload)
	if int64(size) > s.configuration.MaxMessageSize {
		logger.Error(ctx, ErrMessageTooBig, "server.(*Server).validateMessage", slog.Int("size", size))

		return ErrMessageTooBig
	}
	return ctx.Err()
}

/*
SendToAllAnonymous передаёт текстовое сообщение в буферы всех текущих соединений без получателя.

При отсутствии соединений пропускает отправку без ошибки.
При переполнении буфера отключает соединение.

IMPORTANT: После вызова нельзя изменять содержимое payload.
*/
func (s *Server) SendToAllAnonymous(ctx context.Context, payload []byte) error {
	if err := s.validateMessage(ctx, payload); err != nil {
		return err
	}
	return s.anonymousConnectionShard.send(ctx, "", payload)
}

/*
Publish передаёт текстовое сообщение в буферы текущих подписчиков канала.
Сохраняет порядок публикаций внутри канала.

При отсутствии подписчиков пропускает отправку без ошибки.
При переполнении буфера отключает соединение.

IMPORTANT: После вызова нельзя изменять содержимое payload.
*/
func (s *Server) Publish(ctx context.Context, channel string, payload []byte) error {
	if err := s.validateMessage(ctx, payload); err != nil {
		return err
	}
	return s.subscriptionShard(channel).send(ctx, channel, payload)
}

func (s *Server) subscriptionShard(channel string) *registryShard {
	return &s.subscriptionShards[s.shardIndex(channel)]
}

/*
Unsubscribe отписывает текущие соединения получателя от канала.
Возвращает соединения, у которых снята подписка.

Уже помещённые в буфер сообщения не отзывает.
*/
func (s *Server) Unsubscribe(recipient, channel string) []*Connection {
	connections := s.connectionShard(recipient).connectionList(recipient)
	unsubscribedConnections := connections[:0]
	for _, connection := range connections {
		if connection.Unsubscribe(channel) {
			unsubscribedConnections = append(unsubscribedConnections, connection)
		}
	}
	return unsubscribedConnections
}

/*
UnsubscribeByPrefix отписывает текущие соединения получателя от каналов с указанным префиксом.
Возвращает соединения, у которых снята подписка, сгруппированные по каналам.

Уже помещённые в буфер сообщения не отзывает.
*/
func (s *Server) UnsubscribeByPrefix(ctx context.Context, recipient, channelPrefix string) (map[string][]*Connection, error) {
	if channelPrefix == "" {
		logger.Error(ctx, ErrEmptyChannelPrefix, "server.(*Server).UnsubscribeByPrefix", slog.String("recipient", recipient))

		return nil, ErrEmptyChannelPrefix
	}

	connections := s.connectionShard(recipient).connectionList(recipient)
	var unsubscribedConnections map[string][]*Connection
	for _, connection := range connections {
		for _, channel := range connection.unsubscribeByPrefix(channelPrefix) {
			if unsubscribedConnections == nil {
				unsubscribedConnections = make(map[string][]*Connection)
			}
			unsubscribedConnections[channel] = append(unsubscribedConnections[channel], connection)
		}
	}
	return unsubscribedConnections, nil
}

/*
Disconnect закрывает текущие соединения получателя и ожидает завершения их обработчиков.

Не запрещает последующие подключения получателя.
*/
func (s *Server) Disconnect(ctx context.Context, recipient string) error {
	return s.connectionShard(recipient).disconnect(ctx, recipient)
}

/*
DisconnectAllAnonymous закрывает все текущие соединения без получателя и ожидает завершения их обработчиков.

Не запрещает последующие подключения без получателя.
*/
func (s *Server) DisconnectAllAnonymous(ctx context.Context) error {
	return s.anonymousConnectionShard.disconnect(ctx, "")
}

// Shutdown запрещает новые соединения, закрывает существующие и ожидает завершения их обработчиков.
func (s *Server) Shutdown(ctx context.Context) error {
	for i := range s.connectionShards {
		s.connectionShards[i].close()
	}
	s.anonymousConnectionShard.close()
	for i := range s.subscriptionShards {
		s.subscriptionShards[i].mu.Lock()
		s.subscriptionShards[i].isClosed = true
		s.subscriptionShards[i].mu.Unlock()
	}
	doneChan := make(chan struct{})

	go s.waitShutdown(doneChan)

	select {
	case <-doneChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) waitShutdown(doneChan chan<- struct{}) {
	s.wg.Wait()
	close(doneChan)
}

func New(configuration Configuration) *Server {
	s := &Server{
		configuration: configuration,
		seed:          maphash.MakeSeed(),
	}
	for i := range s.connectionShards {
		s.connectionShards[i].connections = make(map[string]map[*Connection]struct{})
	}
	s.anonymousConnectionShard.connections = make(map[string]map[*Connection]struct{})
	for i := range s.subscriptionShards {
		s.subscriptionShards[i].connections = make(map[string]map[*Connection]struct{})
	}
	return s
}
