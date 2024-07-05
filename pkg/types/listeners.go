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
