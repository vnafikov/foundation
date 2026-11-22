package protocol

const (
	EventSubscribed   Event = "subscribed"
	EventUnsubscribed Event = "unsubscribed"
	EventCreated      Event = "created"
	EventUpdated      Event = "updated"
	EventDeleted      Event = "deleted"
	EventNotification Event = "notification"
)

type Event string
