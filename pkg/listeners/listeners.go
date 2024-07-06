package listeners

import (
	"errors"
	"github.com/quix-labs/flash/pkg/types"
	"sync"
)

func NewListener(config *types.ListenerConfig) *Listener {
	if config == nil {
		panic("config is nil")
	}

	return &Listener{
		Config:    config,
		callbacks: make(map[*types.EventCallback]types.Event),
	}
}

type CreateEventCallback func(event types.Event) error
type DeleteEventCallback func(event types.Event) error
type Listener struct {
	Config *types.ListenerConfig

	// Internals
	sync.Mutex
	callbacks      map[*types.EventCallback]types.Event
	listenedEvents types.Event // Use bitwise comparison to check for listened events

	// Trigger client
	_clientCreateEventCallback CreateEventCallback
	_clientDeleteEventCallback DeleteEventCallback
	_clientInitialized         bool
}

/* Callback management */

func (l *Listener) On(event types.Event, callback types.EventCallback) (func() error, error) {
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

func (l *Listener) Dispatch(event *types.ReceivedEvent) {
	for mask := types.Event(1); mask != 0 && mask <= types.EventsAll; mask <<= 1 {
		if event.Event&mask == 0 {
			continue
		}

		if l.listenedEvents&mask == 0 {
			continue
		}

		for callback, listens := range l.callbacks {
			if listens&mask > 0 {
				(*callback)(event)
			}
		}
	}
}

// Init emit all event for first boot */
func (l *Listener) Init(_createCallback CreateEventCallback, _deleteCallback DeleteEventCallback) error {
	l.Lock()
	defer l.Unlock()

	l._clientCreateEventCallback = _createCallback
	l._clientDeleteEventCallback = _deleteCallback

	// Emit all events for initialization
	for targetEvent := types.Event(1); targetEvent != 0 && targetEvent <= types.EventsAll; targetEvent <<= 1 {
		if l.listenedEvents&targetEvent == 0 {
			continue
		}
		if err := _createCallback(targetEvent); err != nil {
			return err
		}
	}

	l._clientInitialized = true
	return nil
}

func (l *Listener) Close() error {

	if !l._clientInitialized {
		return nil
	}

	// Emit all events for initialization
	for targetEvent := types.Event(1); targetEvent != 0 && targetEvent <= types.EventsAll; targetEvent <<= 1 {
		if l.listenedEvents&targetEvent == 0 {
			continue
		}
		if err := l._clientDeleteEventCallback(targetEvent); err != nil {
			return err
		}
	}

	l._clientInitialized = false
	return nil
}

func (l *Listener) addListenedEventIfNeeded(event types.Event) error {

	initialEvents := l.listenedEvents
	l.listenedEvents |= event

	// Trigger event if change appears
	diff := initialEvents ^ l.listenedEvents
	if diff == 0 {
		return nil
	}

	for targetEvent := types.Event(1); targetEvent != 0 && targetEvent <= types.EventsAll; targetEvent <<= 1 {
		if !l._clientInitialized || targetEvent&diff == 0 || targetEvent&event == 0 {
			continue
		}
		l.Lock()
		if err := l._clientCreateEventCallback(targetEvent); err != nil {
			return err
		}
		l.Unlock()
	}

	return nil
}

func (l *Listener) removeListenedEventIfNeeded(event types.Event) error {

	for targetEvent := types.Event(1); targetEvent != 0 && targetEvent <= event; targetEvent <<= 1 {
		if targetEvent&l.listenedEvents == 0 {
			continue
		}
		if l.hasListenersForEvent(targetEvent) {
			continue
		}

		l.listenedEvents &= ^targetEvent
		if l._clientInitialized {
			l.Lock()
			if err := l._clientDeleteEventCallback(targetEvent); err != nil {
				return err
			}
			l.Unlock()
		}
	}
	return nil
}

func (l *Listener) hasListenersForEvent(event types.Event) bool {
	for _, listens := range l.callbacks {
		if listens&event > 0 {
			return true
		}
	}
	return false
}
