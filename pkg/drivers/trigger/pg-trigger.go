package trigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/quix-labs/flash/pkg/types"
)

type DriverConfig struct {
	Schema string // The schema name, which should be unique across all instances
}

func NewDriver(config *DriverConfig) *Driver {
	if config == nil {
		config = &DriverConfig{}
	}
	if config.Schema == "" {
		config.Schema = "flash"
	}
	return &Driver{
		Config:       config,
		activeEvents: make(map[string]bool),
	}
}

type Driver struct {
	Config *DriverConfig

	conn       *pgx.Conn
	listenConn *pgx.Conn
	subChan    chan string
	unsubChan  chan string

	activeEvents map[string]bool

	_clientConfig *types.ClientConfig
}

func (d *Driver) HandleEventListenStart(listenerUid string, lc *types.ListenerConfig, event *types.Event) error {
	createTriggerSql, eventName, err := d.getCreateTriggerSqlForEvent(listenerUid, lc, event)
	if err != nil {
		return err
	}
	_, err = d.sqlExec(d.conn, createTriggerSql)
	if err != nil {
		return err
	}

	return d.addEventToListened(eventName)
}

func (d *Driver) HandleEventListenStop(listenerUid string, lc *types.ListenerConfig, event *types.Event) error {
	createTriggerSql, eventName, err := d.getDeleteTriggerSqlForEvent(listenerUid, lc, event)
	if err != nil {
		return err
	}
	_, err = d.sqlExec(d.conn, createTriggerSql)
	if err != nil {
		return err
	}

	return d.removeEventToListened(eventName)
}

func (d *Driver) Init(_clientConfig *types.ClientConfig) error {
	d._clientConfig = _clientConfig

	var err error
	if d.conn, err = pgx.Connect(context.TODO(), d._clientConfig.DatabaseCnx); err != nil {
		return err
	}

	// Create schema if not exists
	if _, err := d.sqlExec(d.conn, "CREATE SCHEMA IF NOT EXISTS "+d.Config.Schema+";"); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Listen(eventsChan *types.DatabaseEventsChan) error {
	var err error
	if d.listenConn, err = pgx.Connect(context.TODO(), d._clientConfig.DatabaseCnx); err != nil {
		return err
	}

	d.subChan = make(chan string, len(d.activeEvents))
	d.unsubChan = make(chan string, 1)

	// Initialize subChan with activeEvents in queue
	for eventName := range d.activeEvents {
		d.subChan <- eventName
	}

	// Needed because cannot execute LISTEN, UNLISTEN when WaitForNotification is running
	pausedChan := make(chan bool, 1)
	resumeChan := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		for {
			select {
			case eventName := <-d.subChan:
				cancel()
				<-pausedChan // Wait for paused confirmation
				_, err = d.sqlExec(d.listenConn, fmt.Sprintf(`LISTEN "%s"`, eventName))
				if err != nil {
					errChan <- err
					return
				}
				ctx, cancel = context.WithCancel(context.Background()) // Recreate unclosed context
				resumeChan <- true

			case eventName := <-d.unsubChan:
				cancel()
				<-pausedChan // Wait for paused confirmation
				_, err = d.sqlExec(d.listenConn, fmt.Sprintf(`UNLISTEN "%s"`, eventName))
				if err != nil {
					errChan <- err
					return
				}
				ctx, cancel = context.WithCancel(context.Background()) // Recreate unclosed context
				resumeChan <- true
			}
		}
	}()

	go func() {
		for {
			receivedEvent, err := d.listenConn.WaitForNotification(ctx)
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					pausedChan <- true
					<-resumeChan // Wait for resume signal before restart
				} else {
					errChan <- err
				}
				continue
			}

			listenerUid, event, _ := d.parseEventName(receivedEvent.Channel)
			if err != nil {
				errChan <- err
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(receivedEvent.Payload), &data); err != nil {
				errChan <- err
			}

			*eventsChan <- &types.DatabaseEvent{
				ListenerUid: listenerUid,
				ReceivedEvent: &types.ReceivedEvent{
					Event: event,
					Data:  data,
				},
			}
		}
	}()

	return <-errChan
}

func (d *Driver) addEventToListened(eventName string) error {
	d.activeEvents[eventName] = true

	if d.listenConn == nil {
		return nil
	}

	d.subChan <- eventName

	return nil
}

func (d *Driver) removeEventToListened(eventName string) error {
	delete(d.activeEvents, eventName)

	if d.listenConn == nil {
		return nil
	}
	d.unsubChan <- eventName

	return nil
}

func (d *Driver) Close() error {
	// Drop created schema
	if _, err := d.sqlExec(d.conn, "DROP SCHEMA IF EXISTS "+d.Config.Schema+";"); err != nil {
		return err
	}

	// Close active connection
	if d.conn != nil {
		if err := d.conn.Close(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

var (
	_ types.Driver = (*Driver)(nil)
)
