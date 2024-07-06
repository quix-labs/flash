package types

type Driver interface {
	Init(clientConfig *ClientConfig) error
	Close() error

	HandleEventListenStart(listenerUid string, listenerConfig *ListenerConfig, event *Event) error
	HandleEventListenStop(listenerUid string, listenerConfig *ListenerConfig, event *Event) error
	Listen(eventsChan *DatabaseEventsChan) error
}
