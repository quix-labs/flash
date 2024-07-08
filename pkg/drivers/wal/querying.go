package wal

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/quix-labs/flash/pkg/types"
)

type subscription struct {
	listenerUid    string
	listenerConfig *types.ListenerConfig
	event          *types.Event
}

func (d *Driver) initQuerying() error {
	// Bootstrap/Start listening
	d.subChan = make(chan *subscription, 1)
	d.unsubChan = make(chan *subscription, 1)
	d.activePublications = make(map[string]bool)

	return nil
}

func (d *Driver) startQuerying() error {
	// Create connection
	config, err := pgconn.ParseConfig(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}
	config.RuntimeParams["application_name"] = "Flash: replication (querying)"
	if d.queryConn, err = pgconn.ConnectConfig(context.Background(), config); err != nil {
		return err
	}

	// Create false publication to avoid START_REPLICATION error
	initSlotName := d.getFullSlotName("init")
	rawSql, err := d.getCreatePublicationSlotSql(initSlotName, nil, nil)
	if err != nil {
		return err
	}
	if _, err := d.sqlExec(d.queryConn, rawSql); err != nil {
		return err
	}
	d.activePublications[initSlotName] = true

	for {
		select {

		case sub := <-d.unsubChan:
			operationName, err := d.getOperationNameForEvent(sub.event)
			if err != nil {
				return err
			}

			slotName := d.getFullSlotName(sub.listenerUid + "-" + operationName)
			if _, err := d.sqlExec(d.queryConn, d.getDropPublicationSlotSql(slotName)); err != nil {
				return err
			}
			delete(d.activePublications, slotName)
			d.replicationState.restartChan <- struct{}{} // Send restart signal

		case sub := <-d.subChan:
			//TODO USE SAME PUBLICATION USING ALTER
			operationName, err := d.getOperationNameForEvent(sub.event)
			if err != nil {
				return err
			}

			slotName := d.getFullSlotName(sub.listenerUid + "-" + operationName)
			rawSql, err := d.getCreatePublicationSlotSql(slotName, sub.listenerConfig, sub.event)
			if err != nil {
				return err
			}
			if _, err := d.sqlExec(d.queryConn, rawSql); err != nil {
				return err
			}
			d.activePublications[slotName] = true
			d.replicationState.restartChan <- struct{}{} // Send restart signal
		}
	}
}

func (d *Driver) closeQuerying() error {
	for publication, _ := range d.activePublications {
		if _, err := d.sqlExec(d.queryConn, d.getDropPublicationSlotSql(publication)); err != nil {
			return err
		}
	}

	if d.queryConn != nil {
		err := d.queryConn.Close(context.TODO())
		if err != nil {
			return err
		}
	}
	return nil
}
