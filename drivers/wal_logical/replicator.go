package wal_logical

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
	"time"
)

type replicationState struct {
	lastReceivedLSN       pglogrepl.LSN
	currentTransactionLSN pglogrepl.LSN
	lastWrittenLSN        pglogrepl.LSN

	typeMap   *pgtype.Map
	relations map[uint32]*pglogrepl.RelationMessageV2

	processMessages bool
	inStream        bool
	streamQueues    map[uint32][]*pglogrepl.Message

	restartChan chan struct{}
}

func (d *Driver) initReplicator() error {
	d.replicationState = &replicationState{
		lastWrittenLSN: pglogrepl.LSN(0), //TODO KEEP IN FILE OR IGNORE
		relations:      make(map[uint32]*pglogrepl.RelationMessageV2),
		typeMap:        pgtype.NewMap(),
		streamQueues:   make(map[uint32][]*pglogrepl.Message),
		restartChan:    make(chan struct{}),
	}
	return nil
}

func (d *Driver) startReplicator() error {
	if err := d.startConn(); err != nil {
		d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
		return err
	}
	if err := d.startReplication(); err != nil {
		d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
		return err
	}

	/* LISTENING */
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)

	for {
		select {
		case <-d.replicationState.restartChan:

			if d.replicationConn == nil {
				continue
			}
			if err := d.replicationConn.Close(context.Background()); err != nil {
				d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
				return err
			}
			if err := d.startConn(); err != nil {
				d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
				return err
			}
			if err := d.startReplication(); err != nil {
				d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
				return err
			}

			continue

		default:
			if d.replicationConn == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			if time.Now().After(nextStandbyMessageDeadline) && d.replicationState.lastReceivedLSN > 0 {
				err := pglogrepl.SendStandbyStatusUpdate(context.Background(), d.replicationConn, pglogrepl.StandbyStatusUpdate{
					WALWritePosition: d.replicationState.lastWrittenLSN + 1,
					WALFlushPosition: d.replicationState.lastWrittenLSN + 1,
					WALApplyPosition: d.replicationState.lastReceivedLSN + 1,
				})
				if err != nil {
					d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
					return err
				}
				d._clientConfig.Logger.Trace().Msg("Sent Standby status message at " + (d.replicationState.lastWrittenLSN + 1).String())
				nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
			}

			ctx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
			rawMsg, err := d.replicationConn.ReceiveMessage(ctx)
			cancel()

			if err != nil {
				if pgconn.Timeout(err) {
					continue
				}
				d._clientConfig.Logger.Warn().Err(err).Msgf("received err: %s", err)
				time.Sleep(time.Millisecond * 100)
				continue // CLOSED CONNECTION TODO handle and return err when needed
			}

			if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
				return errors.New(errMsg.Message)
			}

			msg, ok := rawMsg.(*pgproto3.CopyData)
			if !ok {
				d._clientConfig.Logger.Warn().Msg(fmt.Sprintf("Received unexpected message: %T", rawMsg))
				continue
			}

			switch msg.Data[0] {
			case pglogrepl.PrimaryKeepaliveMessageByteID:
				pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
				if err != nil {
					return err
				}
				d._clientConfig.Logger.Trace().Msg(fmt.Sprintf("Primary Keepalive Message => ServerWALEnd: %s ServerTime: %s ReplyRequested: %t", pkm.ServerWALEnd, pkm.ServerTime, pkm.ReplyRequested))

				d.replicationState.lastReceivedLSN = pkm.ServerWALEnd

				if pkm.ReplyRequested {
					nextStandbyMessageDeadline = time.Time{}
				}

			case pglogrepl.XLogDataByteID:
				xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
				if err != nil {
					return err
				}
				//d._clientConfig.Logger.Trace().Msg(fmt.Sprintf("XLogData => WALStart %s ServerWALEnd %s ServerTime %s WALData: %s", xld.WALStart, xld.ServerWALEnd, xld.ServerTime, rawMsg))

				updateLsn, err := d.processXld(&xld)
				if err != nil {
					return err
				}
				if updateLsn {
					d.replicationState.lastWrittenLSN = xld.ServerWALEnd
					// TODO write wal position in file if needed
					nextStandbyMessageDeadline = time.Time{} // Force resend standby message
				}
			}

		}
	}
}

func (d *Driver) closeReplicator() error {
	if d.replicationConn != nil {
		// CLOSE ACTUAL
		if err := d.replicationConn.Close(context.Background()); err != nil {
			d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
			return err
		}
		//REMAKE NEW CONN WITHOUT STARTING REPLICATION
		if err := d.startConn(); err != nil {
			d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
			return err
		}
		dropReplicationSql := fmt.Sprintf(`select pg_drop_replication_slot(slot_name) from pg_replication_slots where slot_name = '%s';`, d.Config.ReplicationSlot)
		_, err := d.sqlExec(d.replicationConn, dropReplicationSql)
		if err != nil {
			d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
			return err
		}
		// CLOSE TEMP
		if err := d.replicationConn.Close(context.Background()); err != nil {
			d._clientConfig.Logger.Error().Err(err).Msgf("received err: %s", err)
			return err
		}

		d.replicationConn = nil
	}
	return nil
}

func (d *Driver) startConn() error {
	// Create querying and listening connections
	config, err := pgconn.ParseConfig(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}
	config.RuntimeParams["application_name"] = "Flash: replication (replicator)"
	config.RuntimeParams["replication"] = "database"

	if d.replicationConn, err = pgconn.ConnectConfig(context.Background(), config); err != nil {
		return err
	}

	// Create false publication to avoid START_REPLICATION error
	initSlotName := d.getFullSlotName("init")

	// DROP OLD
	dropPublicationSql := d.getDropPublicationSlotSql(initSlotName)
	dropReplicationSql := fmt.Sprintf(`select pg_drop_replication_slot(slot_name) from pg_replication_slots where slot_name = '%s';`, d.Config.ReplicationSlot)
	createPublicationSlotSql, err := d.getCreatePublicationSlotSql(initSlotName, nil, nil)
	if err != nil {
		return err
	}

	d.activePublications[initSlotName] = true

	if _, err := d.sqlExec(d.replicationConn, dropPublicationSql+dropReplicationSql+createPublicationSlotSql); err != nil {
		return err
	}

	return nil
}

func (d *Driver) startReplication() error {
	if _, err := d.sqlExec(d.replicationConn, fmt.Sprintf(`CREATE_REPLICATION_SLOT "%s" TEMPORARY LOGICAL "pgoutput";`, d.Config.ReplicationSlot)); err != nil {
		return err
	}

	initSlotName := d.getFullSlotName("init")
	activePublications := []string{initSlotName}
	for publicationName, _ := range d.activePublications {
		activePublications = append(activePublications, publicationName)
	}
	replicationOptions := pglogrepl.StartReplicationOptions{
		Mode: pglogrepl.LogicalReplication,
		PluginArgs: []string{
			"proto_version '2'", // Keep as version 2 to compatibility
			"publication_names '" + strings.Join(activePublications, ", ") + "'",
			"messages 'true'",
		},
	}
	if d.Config.UseStreaming {
		replicationOptions.PluginArgs = append(replicationOptions.PluginArgs, "streaming 'true'")
	}

	if err := pglogrepl.StartReplication(context.Background(), d.replicationConn, d.Config.ReplicationSlot, d.replicationState.lastWrittenLSN+1, replicationOptions); err != nil {
		return err
	}
	d._clientConfig.Logger.Debug().Msg("Started replication slot: " + d.Config.ReplicationSlot)
	return nil
}
