package trigger

import (
	"context"
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
		Config: config,
	}
}

type Driver struct {
	Config *DriverConfig

	conn *pgx.Conn

	_clientConfig *types.ClientConfig
}

func (d *Driver) Init(_clientConfig *types.ClientConfig) error {
	d._clientConfig = _clientConfig

	var err error
	if d.conn, err = pgx.Connect(context.TODO(), d._clientConfig.DatabaseCnx); err != nil {
		return err
	}

	// Create schema if not exists
	if _, err := d.conn.Exec(context.TODO(), "CREATE SCHEMA IF NOT EXISTS "+d.Config.Schema+";"); err != nil {
		return err
	}

	return nil
}

func (d *Driver) HandleEventListenStart(lc *types.ListenerConfig, event *types.Event) error {
	createTriggerSql, err := d.getCreateTriggerSqlForEvent(lc, event)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec(context.TODO(), createTriggerSql)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) HandleEventListenStop(lc *types.ListenerConfig, event *types.Event) error {
	createTriggerSql, err := d.getDeleteTriggerSqlForEvent(lc, event)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec(context.TODO(), createTriggerSql)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) Close() error {
	// Drop created schema
	if _, err := d.conn.Exec(context.TODO(), "DROP SCHEMA IF EXISTS "+d.Config.Schema+";"); err != nil {
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
