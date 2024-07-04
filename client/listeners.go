package client

import (
	"errors"
)

type Event uint8

const (
	EventInsert Event = 1 << iota
	EventUpdate
	EventDelete
	EventTruncate

	EventsAll = EventInsert | EventUpdate | EventDelete | EventTruncate
)

type EventCallback func(Event)

type ListenerConfig struct {
	Table string
}

func NewListener(config *ListenerConfig) *Listener {
	if config == nil {
		panic("config is nil")
	}

	return &Listener{
		Config:    config,
		callbacks: make(map[*EventCallback]Event),
	}
}

type CreateEventCallback func(event Event) error
type DeleteEventCallback func(event Event) error
type Listener struct {
	Config *ListenerConfig

	// Internals
	callbacks      map[*EventCallback]Event
	listenedEvents Event // Use bitwise comparison to check for listened events

	// Trigger client
	_clientCreateEventCallback CreateEventCallback
	_clientDeleteEventCallback DeleteEventCallback
	_clientInitialized         bool
}

/* Callback management */

func (l *Listener) On(event Event, callback EventCallback) (func() error, error) {
	if callback == nil {
		return nil, errors.New("callback cannot be nil")
	}

	if err := l.addListenedEventIfNeeded(event); err != nil {
		return nil, err
	}

	l.callbacks[&callback] = event

	removeFunc := func() error {
		delete(l.callbacks, &callback) // Important keep before removeListenedEventIfNeeded
		if err := l.removeListenedEventIfNeeded(event); err != nil {
			return err
		}
		return nil
	}

	return removeFunc, nil
}

func (l *Listener) Dispatch(event Event) {
	for mask := Event(1); mask != 0 && mask <= EventsAll; mask <<= 1 {
		if event&mask == 0 {
			continue
		}

		if l.listenedEvents&mask == 0 {
			continue
		}

		for callback, listens := range l.callbacks {
			if listens&mask > 0 {
				(*callback)(mask)
			}
		}
	}
}

// Init emit all event for first boot */
func (l *Listener) Init(_createCallback CreateEventCallback, _deleteCallback DeleteEventCallback) error {
	//TODO LOCK
	l._clientCreateEventCallback = _createCallback
	l._clientDeleteEventCallback = _deleteCallback

	// Emit all events for initialization
	for mask := Event(1); mask != 0 && mask <= EventsAll; mask <<= 1 {
		if l.listenedEvents&mask == 0 {
			continue
		}
		if err := _createCallback(mask); err != nil {
			return err
		}
	}

	l._clientInitialized = true
	//TODO UNLOCK
	return nil
}

func (l *Listener) addListenedEventIfNeeded(event Event) error {
	initialEvents := l.listenedEvents
	l.listenedEvents |= event

	// Trigger event if change appears
	diff := initialEvents ^ l.listenedEvents
	if diff == 0 {
		return nil
	}

	for targetEvent := Event(1); targetEvent != 0 && targetEvent <= EventsAll; targetEvent <<= 1 {
		if !l._clientInitialized || targetEvent&diff == 0 || targetEvent&event == 0 {
			continue
		}
		if err := l._clientCreateEventCallback(targetEvent); err != nil {
			return err
		}
	}

	return nil
}

func (l *Listener) removeListenedEventIfNeeded(event Event) error {
	for targetEvent := Event(1); targetEvent != 0 && targetEvent <= event; targetEvent <<= 1 {
		if targetEvent&l.listenedEvents == 0 {
			continue
		}
		if l.hasListenersForEvent(targetEvent) {
			continue
		}

		l.listenedEvents &= ^targetEvent
		if l._clientInitialized {
			if err := l._clientDeleteEventCallback(targetEvent); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Listener) hasListenersForEvent(event Event) bool {
	for _, listens := range l.callbacks {
		if listens&event > 0 {
			return true
		}
	}
	return false
}
