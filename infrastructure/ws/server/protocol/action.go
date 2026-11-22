package protocol

const (
	ActionSubscribe   Action = "subscribe"
	ActionUnsubscribe Action = "unsubscribe"
)

var Actions = [...]Action{
	ActionSubscribe,
	ActionUnsubscribe,
}

type Action string
