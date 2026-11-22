package handlers

import (
	"context"

	"github.com/vnafikov/foundation/adapters/ws/server/response"
	"github.com/vnafikov/foundation/core/contracterrors"
	"github.com/vnafikov/foundation/core/parameters"
	"github.com/vnafikov/foundation/infrastructure/ws/server"
	"github.com/vnafikov/foundation/infrastructure/ws/server/presenters"
	"github.com/vnafikov/foundation/infrastructure/ws/server/protocol"
	"github.com/vnafikov/foundation/infrastructure/ws/server/response/render"
	"github.com/vnafikov/foundation/json"
)

type Message struct {
	checkSubscriptionAccess func(ctx context.Context, recipient, channel string) error
}

// Handle обрабатывает подписку и отписку соединения от канала.
func (handler Message) Handle(ctx context.Context, connection *server.Connection, payload []byte) error {
	params := new(messageParameters)
	if err := json.DeserializeInput(ctx, payload, params); err != nil {
		return response.SendError(ctx, connection, "", contracterrors.ErrInvalid)
	}

	if err := parameters.Validate(ctx, params); err != nil {
		channel := ""
		if len(params.Channel) <= protocol.ChannelMaxLen {
			channel = params.Channel
		}
		return response.SendError(ctx, connection, channel, err)
	}

	switch params.Action {
	case protocol.ActionSubscribe:
		return handler.subscribe(ctx, connection, params.Channel)
	case protocol.ActionUnsubscribe:
		return handler.unsubscribe(ctx, connection, params.Channel)
	}
	return response.SendError(ctx, connection, params.Channel, contracterrors.ErrInvalid)
}

func (handler Message) subscribe(ctx context.Context, connection *server.Connection, channel string) error {
	if err := handler.checkSubscriptionAccess(ctx, connection.Recipient(), channel); err != nil {
		return response.SendError(ctx, connection, channel, err)
	}

	if err := connection.Subscribe(channel); err != nil {
		return err
	}
	return render.SendJSON(ctx, connection, presenters.NewSubscribedResponse(channel))
}

func (Message) unsubscribe(ctx context.Context, connection *server.Connection, channel string) error {
	if !connection.Unsubscribe(channel) {
		return response.SendError(ctx, connection, channel, contracterrors.ErrNotFound)
	}
	return render.SendJSON(ctx, connection, presenters.NewUnsubscribedResponse(channel))
}

func NewMessage(checkSubscriptionAccess func(ctx context.Context, recipient, channel string) error) Message {
	return Message{checkSubscriptionAccess: checkSubscriptionAccess}
}
