package trigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/quix-labs/flash/pkg/types"
	"sync"
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

	conn         *pgx.Conn
	listenConn   *pgx.Conn
	subChan      chan string
	unsubChan    chan string
	shutdownChan chan bool

	activeEvents  map[string]bool
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

	connConfig, err := pgx.ParseConfig(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}
	connConfig.Config.RuntimeParams["application_name"] = "Flash: " + d.Config.Schema + " (querying)"
	if d.conn, err = pgx.ConnectConfig(context.TODO(), connConfig); err != nil {
		return err
	}

	// Create schema if not exists
	if _, err := d.sqlExec(d.conn, "CREATE SCHEMA IF NOT EXISTS "+d.Config.Schema+";"); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Listen(eventsChan *types.DatabaseEventsChan) error {
	connConfig, err := pgx.ParseConfig(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}
	connConfig.RuntimeParams["application_name"] = "Flash: " + d.Config.Schema + " (listening)"
	if d.listenConn, err = pgx.ConnectConfig(context.TODO(), connConfig); err != nil {
		return err
	}

	d.subChan = make(chan string, len(d.activeEvents))
	d.unsubChan = make(chan string, 1)

	// Initialize subChan with activeEvents in queue
	for eventName := range d.activeEvents {
		d.subChan <- eventName
	}

	listeningStopped := make(chan bool)
	resumeChan := make(chan bool)
	ctx, stopListening := context.WithCancel(context.Background())
	defer stopListening()

	errChan := make(chan error, 1)
	var mu sync.Mutex // Mutex for synchronizing LISTEN/UNLISTEN

	go func() {
		for {
			select {
			case eventName := <-d.subChan:
				mu.Lock()
				stopListening()
				<-listeningStopped
				if _, err = d.sqlExec(d.listenConn, fmt.Sprintf(`LISTEN "%s"`, eventName)); err != nil {
					errChan <- err
				}
				ctx, stopListening = context.WithCancel(context.Background())
				resumeChan <- true
				mu.Unlock()

			case eventName := <-d.unsubChan:
				fmt.Println("CLAIMED")
				mu.Lock()
				fmt.Println("CLAIM PASSED")
				stopListening()
				<-listeningStopped
				fmt.Println("STOPPED PASSED")

				if _, err = d.sqlExec(d.listenConn, fmt.Sprintf(`UNLISTEN "%s"`, eventName)); err != nil {
					errChan <- err
				}
				ctx, stopListening = context.WithCancel(context.Background())
				resumeChan <- true
				fmt.Println("RESUMED PASSED")
				mu.Unlock()
			}
		}
	}()

	d.shutdownChan = make(chan bool)
	go func() {
		for {
			select {
			case <-d.shutdownChan:
				stopListening()
				<-listeningStopped
				return
			default:
				receivedEvent, err := d.listenConn.WaitForNotification(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						listeningStopped <- true
						<-resumeChan
					} else {
						errChan <- err
					}
					continue
				}

				listenerUid, event, err := d.parseEventName(receivedEvent.Channel)
				if err != nil {
					errChan <- err
					continue
				}

				var data types.EventData
				if receivedEvent.Payload != "" {
					data = make(types.EventData)
					if err := json.Unmarshal([]byte(receivedEvent.Payload), &data); err != nil {
						errChan <- err
						continue
					}
				}

				*eventsChan <- &types.DatabaseEvent{
					ListenerUid: listenerUid,
					ReceivedEvent: &types.ReceivedEvent{
						Event: event,
						Data:  &data,
					},
				}
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
	d.shutdownChan <- true
	// Drop created schema
	if _, err := d.sqlExec(d.conn, "DROP SCHEMA IF EXISTS "+d.Config.Schema+" CASCADE;"); err != nil {
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
