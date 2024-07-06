package client

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/drivers/trigger"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
	"github.com/rs/zerolog"
	"os"
	"strings"
	"sync"
)

func NewClient(config *types.ClientConfig) *Client {
	if config == nil {
		config = &types.ClientConfig{}
	}
	if config.DatabaseCnx == "" {
		panic("database connection required") //TODO Error handling
	}
	if config.Driver == nil {
		config.Driver = trigger.NewDriver(nil)
	}
	if config.Logger == nil {
		logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Stack().Timestamp().Logger()
		config.Logger = &logger
	}
	return &Client{
		Config:    config,
		listeners: make(map[string]*listeners.Listener),
	}
}

type Client struct {
	Config    *types.ClientConfig
	listeners map[string]*listeners.Listener
}

func (c *Client) Attach(l *listeners.Listener) {
	listenerUid := c.getUniqueNameForListener(l)
	c.listeners[listenerUid] = l
}

func (c *Client) Start() error {
	err := c.Init()
	if err != nil {
		return err
	}

	eventChan := make(types.DatabaseEventsChan)
	errChan := make(chan error)
	go func() {
		if err := c.Config.Driver.Listen(&eventChan); err != nil {
			errChan <- err
		}
	}()

	for {
		select {
		case receivedEvent := <-eventChan:
			listener, exists := c.listeners[receivedEvent.ListenerUid]
			if !exists {
				return fmt.Errorf("listener %s not found", receivedEvent.ListenerUid) // I think simply can be ignored
			}
			listener.Dispatch(receivedEvent.ReceivedEvent)
		case err := <-errChan:
			return err
		}
	}
}

func (c *Client) Init() error {
	c.Config.Logger.Debug().Msg("Init driver")
	if err := c.Config.Driver.Init(c.Config); err != nil {
		return err
	}
	c.Config.Logger.Debug().Msg("Init listeners")

	// Init listeners (parallel)
	var wg sync.WaitGroup
	for lUid, l := range c.listeners {
		wg.Add(1)

		listenerUid := lUid // Keep intermediate value to avoid conflict between loop iterations
		listener := l       // Keep intermediate value to avoid conflict between loop iterations

		errChan := make(chan error)
		go func() {
			defer wg.Done()
			err := listener.Init(func(event types.Event) error {
				return c.Config.Driver.HandleEventListenStart(listenerUid, listener.Config, &event)
			}, func(event types.Event) error {
				return c.Config.Driver.HandleEventListenStop(listenerUid, listener.Config, &event)
			})
			errChan <- err
		}()
		err := <-errChan
		if err != nil {
			return err
		}
	}
	wg.Wait()

	c.Config.Logger.Debug().Msg("Listener initialized")
	return nil
}

func (c *Client) Close() {
	var wg sync.WaitGroup

	c.Config.Logger.Debug().Msg("Closing listeners")

	// Remove listeners (parallel)
	errChan := make(chan error)
	for _, l := range c.listeners {
		wg.Add(1)
		listener := l // Keep copy to avoid invalid reference due to for loop
		go func() {
			defer wg.Done()
			if err := listener.Close(); err != nil {
				errChan <- err
			}
		}()
	}
	//TODO HANDLE ERROR IN errChan
	wg.Wait()
	c.Config.Logger.Debug().Msg("Listeners closed")

	c.Config.Logger.Debug().Msg("Closing driver")
	_ = c.Config.Driver.Close()
	c.Config.Logger.Debug().Msg("Driver closed")
}

func (c *Client) getUniqueNameForListener(lc *listeners.Listener) string {
	return strings.ReplaceAll(fmt.Sprintf("%p", lc), "0x", "")
}
