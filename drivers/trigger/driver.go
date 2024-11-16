package trigger

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/quix-labs/flash"
	"net/url"
	"time"
)

type DriverConfig struct {
	Schema string // The schema name, which should be unique across all instances
}

var (
	_ flash.Driver = (*Driver)(nil) // Interface implementation
)

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

	conn       *sql.DB
	pgListener *pq.Listener

	subChan   chan string
	unsubChan chan string
	shutdown  chan bool

	activeEvents  map[string]bool
	_clientConfig *flash.ClientConfig
}

func (d *Driver) HandleOperationListenStart(listenerUid string, lc *flash.ListenerConfig, operation flash.Operation) error {
	createTriggerSql, eventName, err := d.getCreateTriggerSqlForOperation(listenerUid, lc, &operation)
	if err != nil {
		return err
	}
	_, err = d.sqlExec(d.conn, createTriggerSql)
	if err != nil {
		return err
	}

	return d.addEventToListened(eventName)
}

func (d *Driver) HandleOperationListenStop(listenerUid string, lc *flash.ListenerConfig, event flash.Operation) error {
	createTriggerSql, eventName, err := d.getDeleteTriggerSqlForEvent(listenerUid, lc, &event)
	if err != nil {
		return err
	}
	_, err = d.sqlExec(d.conn, createTriggerSql)
	if err != nil {
		return err
	}

	return d.removeEventToListened(eventName)
}

func (d *Driver) Init(_clientConfig *flash.ClientConfig) error {
	d._clientConfig = _clientConfig

	parsedCnx, err := url.Parse(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}

	query := parsedCnx.Query()
	query.Set("application_name", "test")
	parsedCnx.RawQuery = query.Encode()

	connector, err := pq.NewConnector(parsedCnx.String())
	if err != nil {
		return err
	}

	d.conn = sql.OpenDB(connector)
	// Create schema if not exists
	if _, err := d.sqlExec(d.conn, "CREATE SCHEMA IF NOT EXISTS \""+d.Config.Schema+"\";"); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Listen(eventsChan *flash.DatabaseEventsChan) error {
	errChan := make(chan error)
	d.subChan = make(chan string, len(d.activeEvents))
	d.unsubChan = make(chan string, 1)
	d.shutdown = make(chan bool)

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			errChan <- err
		}
	}

	parsedCnx, err := url.Parse(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}

	query := parsedCnx.Query()
	query.Set("application_name", "test_listen")
	parsedCnx.RawQuery = query.Encode()

	d.pgListener = pq.NewListener(parsedCnx.String(), 1*time.Second, time.Minute, reportProblem)

	// Initialize subChan with activeEvents in queue
	go func() {
		for eventName := range d.activeEvents {
			d.subChan <- eventName
		}
	}()

	for {
		select {

		case <-d.shutdown:
			return d.pgListener.Close()

		case err := <-errChan:
			return err

		case eventName := <-d.unsubChan:
			d._clientConfig.Logger.Trace().Str("query", fmt.Sprintf(`UNLISTEN "%s"`, eventName)).Msg("sending sql request")
			if err := d.pgListener.Unlisten(eventName); err != nil {
				return err
			}
			continue

		case eventName := <-d.subChan:
			d._clientConfig.Logger.Trace().Str("query", fmt.Sprintf(`LISTEN "%s"`, eventName)).Msg("sending sql request")
			if err := d.pgListener.Listen(eventName); err != nil {
				return err
			}
			continue

		case notification := <-d.pgListener.Notify:
			listenerUid, operation, err := d.parseEventName(notification.Channel)
			if err != nil {
				errChan <- err
				continue
			}

			var data map[string]any
			if notification.Extra != "" {
				data = make(map[string]any)
				if err := json.Unmarshal([]byte(notification.Extra), &data); err != nil {
					errChan <- err
					continue
				}
			}
			var newData, oldData *flash.EventData = nil, nil
			if data != nil {
				if nd, exists := data["new"]; exists && nd != nil {
					typedData := flash.EventData(nd.(map[string]any))
					newData = &typedData
				}
				if od, exists := data["old"]; exists && od != nil {
					typedData := flash.EventData(od.(map[string]any))
					oldData = &typedData
				}
			}

			// Custom conditions if update to handle soft deletes
			if operation == flash.OperationUpdate {
				var previouslyMatch, newlyMatch bool = true, true
				/* Extract condition match */
				if nc, exists := data["new_condition"]; exists && nc != nil {
					newlyMatch = nc.(bool)
				}
				if oc, exists := data["old_condition"]; exists && oc != nil {
					previouslyMatch = oc.(bool)
				}

				// Send insert signal
				if !previouslyMatch && newlyMatch {
					*eventsChan <- &flash.DatabaseEvent{
						ListenerUid: listenerUid,
						Event:       &flash.InsertEvent{New: newData},
					}
				} else if previouslyMatch && !newlyMatch {
					*eventsChan <- &flash.DatabaseEvent{
						ListenerUid: listenerUid,
						Event:       &flash.DeleteEvent{Old: oldData},
					}
				} else if previouslyMatch && newlyMatch {
					*eventsChan <- &flash.DatabaseEvent{
						ListenerUid: listenerUid,
						Event:       &flash.UpdateEvent{New: newData, Old: oldData},
					}
				}
				continue
			}

			switch operation {
			case flash.OperationInsert:
				*eventsChan <- &flash.DatabaseEvent{
					ListenerUid: listenerUid,
					Event:       &flash.InsertEvent{New: newData},
				}
			case flash.OperationUpdate:
				*eventsChan <- &flash.DatabaseEvent{
					ListenerUid: listenerUid,
					Event:       &flash.UpdateEvent{New: newData, Old: oldData},
				}
			case flash.OperationDelete:
				*eventsChan <- &flash.DatabaseEvent{
					ListenerUid: listenerUid,
					Event:       &flash.DeleteEvent{Old: oldData},
				}
			case flash.OperationTruncate:
				*eventsChan <- &flash.DatabaseEvent{
					ListenerUid: listenerUid,
					Event:       &flash.TruncateEvent{},
				}
			default:
				return fmt.Errorf("unknown operation: %d", operation)
			}
		}
	}
}

func (d *Driver) addEventToListened(eventName string) error {
	d.activeEvents[eventName] = true

	if d.pgListener == nil {
		return nil
	}

	d.subChan <- eventName

	return nil
}

func (d *Driver) removeEventToListened(eventName string) error {
	delete(d.activeEvents, eventName)

	if d.pgListener == nil {
		return nil
	}
	d.unsubChan <- eventName

	return nil
}

func (d *Driver) Close() error {
	if d.pgListener != nil {
		d.shutdown <- true
	}

	// Drop created schema
	if _, err := d.sqlExec(d.conn, "DROP SCHEMA IF EXISTS \""+d.Config.Schema+"\" CASCADE;"); err != nil {
		return err
	}

	// Close active connection
	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
