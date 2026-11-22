package services

import (
	"context"
	"errors"

	"github.com/vnafikov/foundation/infrastructure/ws/server"
	"github.com/vnafikov/foundation/infrastructure/ws/server/presenters"
	"github.com/vnafikov/foundation/json"
)

type Subscriptions struct {
	wsServer *server.Server
}

// Unsubscribe отписывает текущие соединения получателя от канала.
// Отправляет клиенту подтверждение снятия подписки.
func (subscriptions Subscriptions) Unsubscribe(ctx context.Context, recipient, channel string) error {
	connections := subscriptions.wsServer.Unsubscribe(recipient, channel)
	return subscriptions.sendUnsubscribed(ctx, connections, channel)
}

func (Subscriptions) sendUnsubscribed(ctx context.Context, connections []*server.Connection, channel string) error {
	if len(connections) == 0 {
		return nil
	}

	payload, err := json.Serialize(ctx, presenters.NewUnsubscribedResponse(channel))
	if err != nil {
		return err
	}

	for _, connection := range connections {
		e := connection.Send(ctx, payload)
		if e != nil && !errors.Is(e, server.ErrConnectionClosed) {
			err = errors.Join(err, e)
		}
	}
	return err
}

// UnsubscribeByPrefix отписывает текущие соединения получателя от каналов с указанным префиксом.
// Отправляет клиенту подтверждения снятия подписок.
func (subscriptions Subscriptions) UnsubscribeByPrefix(ctx context.Context, recipient, channelPrefix string) error {
	unsubscribedConnections, err := subscriptions.wsServer.UnsubscribeByPrefix(ctx, recipient, channelPrefix)
	if err != nil {
		return err
	}

	for channel, connections := range unsubscribedConnections {
		if e := subscriptions.sendUnsubscribed(ctx, connections, channel); e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func NewSubscriptions(wsServer *server.Server) Subscriptions {
	return Subscriptions{wsServer: wsServer}
}
