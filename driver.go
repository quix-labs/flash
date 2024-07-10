package flash

type DatabaseEvent struct {
	ListenerUid string
	Event       Event
}
type DatabaseEventsChan chan *DatabaseEvent
type Driver interface {
	Init(clientConfig *ClientConfig) error
	Close() error

	HandleEventListenStart(listenerUid string, listenerConfig *ListenerConfig, event *Operation) error
	HandleEventListenStop(listenerUid string, listenerConfig *ListenerConfig, event *Operation) error
	Listen(eventsChan *DatabaseEventsChan) error
}
