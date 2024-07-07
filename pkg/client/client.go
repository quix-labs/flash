package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/quix-labs/flash/pkg/drivers/trigger"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
	"github.com/rs/zerolog"
	"os"
	"strings"
	"sync"
	"time"
)

func NewClient(config *types.ClientConfig) (*Client, error) {
	if config == nil {
		config = &types.ClientConfig{}
	}
	if config.DatabaseCnx == "" {
		return nil, errors.New("database connection required")
	}
	if config.Driver == nil {
		config.Driver = trigger.NewDriver(nil)
	}
	if config.Logger == nil {
		logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Stack().Timestamp().Logger()
		config.Logger = &logger
	}
	if config.ShutdownTimeout == time.Duration(0) {
		config.ShutdownTimeout = 10 * time.Second
	}
	return &Client{
		Config:    config,
		listeners: make(map[string]*listeners.Listener),
	}, nil
}

type Client struct {
	Config    *types.ClientConfig
	listeners map[string]*listeners.Listener
}

func (c *Client) Attach(l *listeners.Listener) {
	listenerUid := c.getUniqueNameForListener(l)
	c.listeners[listenerUid] = l
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

func (c *Client) Close() error {
	errChan := make(chan error, 1)
	go func() {
		//TODO PARALLEL
		c.Config.Logger.Debug().Msg("Closing listeners")
		for _, l := range c.listeners {
			if err := l.Close(); err != nil {
				c.Config.Logger.Error().Err(err).Msg("Error closing listener")
				errChan <- err
				return
			}
		}
		c.Config.Logger.Debug().Msg("Listeners closed")

		c.Config.Logger.Debug().Msg("Closing driver")
		errChan <- c.Config.Driver.Close()
	}()

	// Create timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), c.Config.ShutdownTimeout)
	defer cancel()

	select {
	case err := <-errChan:
		if err != nil {
			c.Config.Logger.Error().Err(err).Msg("Failed to close driver")
			return err
		}
		c.Config.Logger.Debug().Msg("Driver closed")

	case <-ctx.Done():
		c.Config.Logger.Error().Msg("timeout reached while closing, some events can be loss")
	}

	return nil
}

func (c *Client) getUniqueNameForListener(lc *listeners.Listener) string {
	return strings.ReplaceAll(fmt.Sprintf("%p", lc), "0x", "")
}
