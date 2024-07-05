package types

type Driver interface {
	Init(*ClientConfig) error
	Close() error

	HandleEventListenStart(*ListenerConfig, *Event) error
	HandleEventListenStop(*ListenerConfig, *Event) error
}
