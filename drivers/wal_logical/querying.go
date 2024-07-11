package wal_logical

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/quix-labs/flash"
)

type subscriptionClaim struct {
	listenerUid    string
	listenerConfig *flash.ListenerConfig
	operation      *flash.Operation
}

type activePublication struct {
	listenerConfig *flash.ListenerConfig
	slotName       string
	operations     *flash.Operation // Use with bitwise to handle combined operations
}

// Key -> listenerUid
type subscriptionState struct {
	subChan              chan *subscriptionClaim
	unsubChan            chan *subscriptionClaim
	currentSubscriptions map[string]*activePublication
}

func (d *Driver) initQuerying() error {
	d.subscriptionState = &subscriptionState{
		subChan:              make(chan *subscriptionClaim),
		unsubChan:            make(chan *subscriptionClaim),
		currentSubscriptions: make(map[string]*activePublication),
	}

	// Bootstrap/Start listening TODO USELESS
	d.activePublications = make(map[string]bool)

	return nil
}

func (d *Driver) startQuerying(readyChan *chan struct{}) error {
	// Create connection
	config, err := pgconn.ParseConfig(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}
	config.RuntimeParams["application_name"] = "Flash: replication (querying)"
	if d.queryConn, err = pgconn.ConnectConfig(context.Background(), config); err != nil {
		return err
	}

	*readyChan <- struct{}{}
	for {
		select {

		case claimSub := <-d.subscriptionState.unsubChan:
			currentSub, exists := d.subscriptionState.currentSubscriptions[claimSub.listenerUid]
			if !exists {
				continue
			}

			// TODO Operation.Remove()
			prevEvents := *currentSub.operations
			*currentSub.operations &= ^(*claimSub.operation) // Remove operation from listened

			// Bypass if no changes
			if *currentSub.operations == prevEvents {
				return nil
			}

			if len(currentSub.operations.GetAtomics()) > 0 {
				alterSql, err := d.getAlterPublicationEventsSql(currentSub)
				if err != nil {
					return err
				}
				if _, err := d.sqlExec(d.queryConn, alterSql); err != nil {
					return err
				}
			} else {
				if _, err := d.sqlExec(d.queryConn, d.getDropPublicationSlotSql(currentSub.slotName)); err != nil {
					return err
				}
				delete(d.activePublications, currentSub.slotName)
				delete(d.subscriptionState.currentSubscriptions, claimSub.listenerUid)
			}

		case claimSub := <-d.subscriptionState.subChan:
			currentSub, exists := d.subscriptionState.currentSubscriptions[claimSub.listenerUid]
			if !exists {
				currentSub = &activePublication{
					listenerConfig: claimSub.listenerConfig,
					slotName:       d.getFullSlotName(claimSub.listenerUid),
					operations:     claimSub.operation,
				}

				slotName := d.getFullSlotName(claimSub.listenerUid)
				rawSql, err := d.getCreatePublicationSlotSql(slotName, claimSub.listenerConfig, claimSub.operation)
				if err != nil {
					return err
				}
				if _, err := d.sqlExec(d.queryConn, rawSql); err != nil {
					return err
				}

				d.subscriptionState.currentSubscriptions[claimSub.listenerUid] = currentSub
				d.activePublications[slotName] = true
				d.replicationState.restartChan <- struct{}{} // Send restart signal

			} else {
				prevEvents := *currentSub.operations

				// TODO Operation.Append() or Operation.Merge()
				*currentSub.operations |= *claimSub.operation //Append operation to listened

				// Bypass if no changes
				if prevEvents == *currentSub.operations {
					return nil
				}

				alterSql, err := d.getAlterPublicationEventsSql(currentSub)
				if err != nil {
					return err
				}
				if _, err := d.sqlExec(d.queryConn, alterSql); err != nil {
					return err
				}
			}
		}
	}
}

func (d *Driver) closeQuerying() error {
	if d.queryConn != nil {
		for publication, _ := range d.activePublications {
			if _, err := d.sqlExec(d.queryConn, d.getDropPublicationSlotSql(publication)); err != nil {
				return err
			}
		}
		err := d.queryConn.Close(context.Background())
		if err != nil {
			return err
		}
		d.queryConn = nil
	}
	return nil
}
