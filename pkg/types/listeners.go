package types

type ListenerConfig struct {
	Table  string   // Can be prefixed by schema - e.g: public.posts
	Fields []string // Empty fields means all ( SELECT * )
}

type Event uint8

const (
	EventInsert Event = 1 << iota
	EventUpdate
	EventDelete
	EventTruncate

	EventsAll = EventInsert | EventUpdate | EventDelete | EventTruncate
)

type EventData map[string]any

type EventCallback func(event *ReceivedEvent)
type ReceivedEvent struct {
	Event Event
	Data  *EventData
}
