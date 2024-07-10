package wal_logical

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/quix-labs/flash"
)

type subscriptionClaim struct {
	listenerUid    string
	listenerConfig *flash.ListenerConfig
	event          *flash.Operation
}

type activePublication struct {
	listenerConfig *flash.ListenerConfig
	slotName       string
	events         *flash.Operation // Use with bitwise to handle combined events
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

		case claimSub := <-d.subscriptionState.unsubChan:
			currentSub, exists := d.subscriptionState.currentSubscriptions[claimSub.listenerUid]
			if !exists {
				continue
			}

			prevEvents := *currentSub.events
			*currentSub.events &= ^(*claimSub.event) // Remove event from listened

			// Bypass if no changes
			if *currentSub.events == prevEvents {
				return nil
			}

			if *currentSub.events > 0 {
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
					events:         claimSub.event,
				}

				slotName := d.getFullSlotName(claimSub.listenerUid)
				rawSql, err := d.getCreatePublicationSlotSql(slotName, claimSub.listenerConfig, claimSub.event)
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
				prevEvents := *currentSub.events
				*currentSub.events |= *claimSub.event //Append event to listened

				// Bypass if no changes
				if prevEvents == *currentSub.events {
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
