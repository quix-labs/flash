package wal

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
	streamQueues    map[uint32][]pglogrepl.Message

	restartChan chan struct{}
}

func (d *Driver) initReplicator() error {
	d.replicationState = &replicationState{
		lastWrittenLSN: pglogrepl.LSN(0), //TODO KEEP IN FILE OR IGNORE
		relations:      make(map[uint32]*pglogrepl.RelationMessageV2),
		typeMap:        pgtype.NewMap(),
		streamQueues:   make(map[uint32][]pglogrepl.Message),
		restartChan:    make(chan struct{}),
	}
	return nil
}

func (d *Driver) startReplicator() error {
	if err := d.restartConn(); err != nil {
		return err
	}

	/* LISTENING */
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)

	for {
		select {
		case <-d.replicationState.restartChan:
			if err := d.restartConn(); err != nil {
				return err
			}
		default:
			if time.Now().After(nextStandbyMessageDeadline) && d.replicationState.lastReceivedLSN > 0 {
				err := pglogrepl.SendStandbyStatusUpdate(context.Background(), d.replicationConn, pglogrepl.StandbyStatusUpdate{
					WALWritePosition: d.replicationState.lastWrittenLSN + 1,
					WALFlushPosition: d.replicationState.lastWrittenLSN + 1,
					WALApplyPosition: d.replicationState.lastReceivedLSN + 1,
				})
				if err != nil {
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
				return err //TODO handle Error With Retry
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

func (d *Driver) restartConn() error {
	if d.replicationConn != nil {
		if err := d.replicationConn.Close(context.Background()); err != nil {
			return err
		}
	}
	// Create querying and listening connections
	config, err := pgconn.ParseConfig(d._clientConfig.DatabaseCnx)
	if err != nil {
		return err
	}
	config.RuntimeParams["application_name"] = "Flash: replication (replicator)"
	config.RuntimeParams["replication"] = "database"
	if d.replicationConn, err = pgconn.ConnectConfig(context.TODO(), config); err != nil {
		return err
	}

	// Get last X Post for starting event
	//sysident, err := pglogrepl.IdentifySystem(context.Background(), d.replicationConn)
	//if err != nil {
	//	return err
	//}
	//
	//clientXLogPos := sysident.XLogPos

	if _, err := d.sqlExec(d.replicationConn, fmt.Sprintf(`CREATE_REPLICATION_SLOT "%s" TEMPORARY LOGICAL "pgoutput";`, d.Config.ReplicationSlot)); err != nil {
		return err
	}
	d._clientConfig.Logger.Debug().Msg("Created temporary replication slot: " + d.Config.ReplicationSlot)

	var activePublications []string
	for publicationName, _ := range d.activePublications {
		activePublications = append(activePublications, publicationName)
	}
	replicationOptions := pglogrepl.StartReplicationOptions{
		Mode: pglogrepl.LogicalReplication,
		PluginArgs: []string{
			"proto_version '4'",
			"publication_names '" + strings.Join(activePublications, ", ") + "'",
			"messages 'true'",
		},
	}
	if d.Config.UseStreaming {
		replicationOptions.PluginArgs = append(replicationOptions.PluginArgs, "streaming 'true'") //TODO Check if parallel can be safely used
	}

	if err := pglogrepl.StartReplication(context.Background(), d.replicationConn, d.Config.ReplicationSlot, d.replicationState.lastWrittenLSN+1, replicationOptions); err != nil {
		return err
	}
	d._clientConfig.Logger.Debug().Msg("Started replication slot: " + d.Config.ReplicationSlot)
	fmt.Println(replicationOptions.PluginArgs)
	return nil
}
