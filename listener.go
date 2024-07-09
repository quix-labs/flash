package flash

import (
	"errors"
	"sync"
)

// TODO SORTIR VERIFICATION AU NIVEAU LISTENER, PBM oblige à envoyer les columns dans l'event
type ListenerCondition struct {
	Column string
	//Operator string //TODO actually only equals are implemented
	Value any
}

type ListenerConfig struct {
	Table              string   // Can be prefixed by schema - e.g: public.posts
	Fields             []string // Empty fields means all ( SELECT * )
	MaxParallelProcess int      // Default to 1 (not parallel) -> use -1 for Infinity

	Conditions []*ListenerCondition
}

type CreateEventCallback func(event Operation) error
type DeleteEventCallback func(event Operation) error

type Listener struct {
	Config *ListenerConfig

	// Internals
	sync.Mutex
	callbacks      map[*EventCallback]Operation
	listenedEvents Operation // Use bitwise comparison to check for listened events
	semaphore      chan struct{}

	// Trigger client
	_clientCreateEventCallback CreateEventCallback
	_clientDeleteEventCallback DeleteEventCallback
	_clientInitialized         bool
}

func NewListener(config *ListenerConfig) (*Listener, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if config.MaxParallelProcess == 0 {
		config.MaxParallelProcess = 1
	}

	var semaphore chan struct{} = nil
	if config.MaxParallelProcess != -1 {
		semaphore = make(chan struct{}, config.MaxParallelProcess)
	}

	return &Listener{
		Config:    config,
		callbacks: make(map[*EventCallback]Operation),
		semaphore: semaphore,
	}, nil
}

/* Callback management */

func (l *Listener) On(operation Operation, callback EventCallback) (func() error, error) {
	if callback == nil {
		return nil, errors.New("callback cannot be nil")
	}

	if err := l.addListenedEventIfNeeded(operation); err != nil {
		return nil, err
	}

	l.callbacks[&callback] = operation

	removeFunc := func() error {
		delete(l.callbacks, &callback) // Important keep before removeListenedEventIfNeeded
		if err := l.removeListenedEventIfNeeded(operation); err != nil {
			return err
		}
		callback = nil
		return nil
	}

	return removeFunc, nil
}

func (l *Listener) Dispatch(event *Event) {
	for callback, listens := range l.callbacks {
		if listens&(*event).GetOperation() > 0 {
			if l.Config.MaxParallelProcess == -1 {
				go (*callback)(*event)
				continue
			}

			// Acquire semaphore
			l.semaphore <- struct{}{}
			if l.Config.MaxParallelProcess == 1 {
				(*callback)(*event)
				<-l.semaphore
				continue
			}

			go func() {
				(*callback)(*event)
				<-l.semaphore
			}()
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
	for targetEvent := Operation(1); targetEvent != 0 && targetEvent <= OperationAll; targetEvent <<= 1 {
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

func (l *Listener) addListenedEventIfNeeded(event Operation) error {

	initialEvents := l.listenedEvents
	l.listenedEvents |= event

	// Trigger event if change appears
	diff := initialEvents ^ l.listenedEvents
	if diff == 0 {
		return nil
	}

	for targetEvent := Operation(1); targetEvent != 0 && targetEvent <= OperationAll; targetEvent <<= 1 {
		if targetEvent&diff == 0 || targetEvent&event == 0 {
			continue
		}
		l.Lock()
		if l._clientInitialized {
			if err := l._clientCreateEventCallback(targetEvent); err != nil {
				return err
			}
		}
		l.Unlock()
	}

	return nil
}

func (l *Listener) removeListenedEventIfNeeded(event Operation) error {

	for targetEvent := Operation(1); targetEvent != 0 && targetEvent <= event; targetEvent <<= 1 {
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

func (l *Listener) Close() error {
	l.Lock()
	defer l.Unlock()
	l._clientInitialized = false
	return nil
}
func (l *Listener) hasListenersForEvent(event Operation) bool {
	for _, listens := range l.callbacks {
		if listens&event > 0 {
			return true
		}
	}
	return false
}
