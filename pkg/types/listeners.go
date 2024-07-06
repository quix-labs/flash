package types

type ListenerConfig struct {
	Table string
}

type Event uint8

const (
	EventInsert Event = 1 << iota
	EventUpdate
	EventDelete
	EventTruncate

	EventsAll = EventInsert | EventUpdate | EventDelete | EventTruncate
)

type EventCallback func(event *ReceivedEvent)
type ReceivedEvent struct {
	Event Event
	Data  any //TODO
}
