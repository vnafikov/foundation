package handlers

import (
	"slices"

	"github.com/vnafikov/foundation/infrastructure/ws/server/protocol"
)

type messageParameters struct {
	Action  protocol.Action `json:"action"`
	Channel string          `json:"channel"`
}

func (p *messageParameters) IsValid() bool {
	return p.isValidAction() &&
		p.isValidChannel()
}

func (p *messageParameters) isValidAction() bool {
	return slices.Contains(protocol.Actions[:], p.Action)
}

func (p *messageParameters) isValidChannel() bool {
	return p.Channel != "" &&
		len(p.Channel) <= protocol.ChannelMaxLen
}
