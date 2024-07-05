package client

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/drivers/trigger"
	"github.com/quix-labs/flash/pkg/types"
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

	return &Client{Config: config}
}

type Client struct {
	Config    *types.ClientConfig
	listeners []*Listener
}

func (c *Client) AddListener(l *Listener) {
	c.listeners = append(c.listeners, l)
}

func (c *Client) Start() error {
	err := c.Init()
	if err != nil {
		return err
	}

	// START LISTENING NOTIFY
	return nil
}

func (c *Client) Init() error {

	fmt.Println("Init driver")
	if err := c.Config.Driver.Init(c.Config); err != nil {
		return err
	}

	fmt.Println("Init listeners")
	// Init listeners (parallel)
	var wg sync.WaitGroup
	for _, l := range c.listeners {
		wg.Add(1)
		go func() error {
			defer wg.Done()
			err := l.Init(func(event types.Event) error {
				return c.Config.Driver.HandleEventListenStart(l.Config, &event)
			}, func(event types.Event) error {
				return c.Config.Driver.HandleEventListenStop(l.Config, &event)
			})
			if err != nil {
				return err
			}

			return nil
		}() // TODO Error handling
	}
	wg.Wait()
	fmt.Println("Listeners initialized")
	return nil
}
func (c *Client) Close() {
	var wg sync.WaitGroup

	// Remove listeners (parallel)
	for _, l := range c.listeners {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Close() //TODO ERROR HANDLING
		}()
	}
	wg.Wait()

	c.Config.Driver.Close() //TODO ERROR HANDLING

	fmt.Println("CLOSING")
}
