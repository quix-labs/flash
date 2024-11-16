package flash

type DatabaseEvent struct {
	ListenerUid string
	Event       Event
}
type DatabaseEventsChan chan *DatabaseEvent
type Driver interface {
	Init(clientConfig *ClientConfig) error
	Close() error

	HandleOperationListenStart(listenerUid string, listenerConfig *ListenerConfig, operation Operation) error
	HandleOperationListenStop(listenerUid string, listenerConfig *ListenerConfig, operation Operation) error
	Listen(eventsChan *DatabaseEventsChan) error
}
