package presenters

import "github.com/vnafikov/foundation/infrastructure/ws/server/protocol"

type SubscribedResponse struct {
	Event   protocol.Event `json:"event"`
	Channel string         `json:"channel"`
}

func NewSubscribedResponse(channel string) SubscribedResponse {
	return SubscribedResponse{Event: protocol.EventSubscribed, Channel: channel}
}

type UnsubscribedResponse struct {
	Event   protocol.Event `json:"event"`
	Channel string         `json:"channel"`
}

func NewUnsubscribedResponse(channel string) UnsubscribedResponse {
	return UnsubscribedResponse{Event: protocol.EventUnsubscribed, Channel: channel}
}

type ErrorResponse struct {
	Channel string             `json:"channel,omitzero"`
	Error   protocol.ErrorCode `json:"error"`
}

func NewErrorResponse(channel string, code protocol.ErrorCode) ErrorResponse {
	return ErrorResponse{Channel: channel, Error: code}
}
