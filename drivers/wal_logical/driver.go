package wal_logical

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/quix-labs/flash"
)

type DriverConfig struct {
	PublicationSlotPrefix string // Default to flash_publication -> Must be unique across all your instances
	ReplicationSlot       string // Default to flash_replication -> Must be unique across all your instances
	UseStreaming          bool   // Default to false -> allow usage of stream for big transaction, can have big memory impact
}

var (
	_ flash.Driver = (*Driver)(nil) // Interface implementation
)

func NewDriver(config *DriverConfig) *Driver {
	if config == nil {
		config = &DriverConfig{}
	}
	if config.PublicationSlotPrefix == "" {
		config.PublicationSlotPrefix = "flash_publication"
	}
	if config.ReplicationSlot == "" {
		config.ReplicationSlot = "flash_replication"
	}
	return &Driver{
		Config:          config,
		activeListeners: make(map[string]map[string]*flash.ListenerConfig),
	}
}

// TODO
type PublicationState map[string]*struct {
	listenedEvents  []flash.Operation
	listenerMapping map[flash.Operation]struct {
		_listenerUid *string
		_config      *flash.ListenerConfig
	}
}

type Driver struct {
	Config *DriverConfig

	queryConn *pgconn.PgConn

	// Replication handling
	replicationConn  *pgconn.PgConn
	replicationState *replicationState

	activePublications map[string]bool
	activeListeners    map[string]map[string]*flash.ListenerConfig // key 1: tableName -> key 2: listenerUid
	eventsChan         *flash.DatabaseEventsChan

	subscriptionState *subscriptionState

	_clientConfig *flash.ClientConfig
}

func (d *Driver) Init(clientConfig *flash.ClientConfig) error {
	d._clientConfig = clientConfig

	if err := d.initQuerying(); err != nil {
		return err
	}

	if err := d.initReplicator(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) HandleEventListenStart(listenerUid string, listenerConfig *flash.ListenerConfig, event *flash.Operation) error {
	tableName := d.sanitizeTableName(listenerConfig.Table, false)

	//TODO ALTER PUBLICATION noinsert SET (publish = 'update, delete');
	if _, exists := d.activeListeners[tableName]; !exists {
		d.activeListeners[tableName] = make(map[string]*flash.ListenerConfig)
	}

	// Keep in goroutine because channel is listened on start
	go func() {
		d.subscriptionState.subChan <- &subscriptionClaim{
			listenerUid:    listenerUid,
			listenerConfig: listenerConfig,
			event:          event,
		}
	}()

	d.activeListeners[tableName][listenerUid] = listenerConfig //TODO MORE PERFORMANT STRUCTURE
	return nil
}

func (d *Driver) HandleEventListenStop(listenerUid string, listenerConfig *flash.ListenerConfig, event *flash.Operation) error {
	tableName := d.sanitizeTableName(listenerConfig.Table, false)

	// Keep in goroutine because channel is listened on start
	go func() {
		d.subscriptionState.unsubChan <- &subscriptionClaim{
			listenerUid:    listenerUid,
			listenerConfig: listenerConfig,
			event:          event,
		}
	}()

	delete(d.activeListeners[tableName], listenerUid) //TODO MORE PERFORMANT STRUCTURE
	return nil
}

func (d *Driver) Listen(eventsChan *flash.DatabaseEventsChan) error {
	d.eventsChan = eventsChan

	var errChan = make(chan error)

	go func() {
		if err := d.startQuerying(); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := d.startReplicator(); err != nil {
			errChan <- err
		}
	}()

	for {
		select {
		case err := <-errChan:
			return err
		}
	}
}

func (d *Driver) Close() error {
	return d.closeQuerying()
}
