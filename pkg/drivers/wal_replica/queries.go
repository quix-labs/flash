package wal_replica

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/quix-labs/flash/pkg/types"
	"strings"
)

func (d *Driver) getFullSlotName(slotName string) string {
	return d.Config.PublicationSlotPrefix + "-" + slotName
}

func (d *Driver) getCreatePublicationSlotSql(fullSlotName string, config *types.ListenerConfig, event *types.Operation) (string, error) {
	if config == nil {
		return fmt.Sprintf(`CREATE PUBLICATION "%s";`, fullSlotName), nil
	}

	// SET REPLICA IDENTITY TO FULL ON CREATION
	quotedTableName := d.sanitizeTableName(config.Table, true)
	rawSql := fmt.Sprintf(`ALTER TABLE %s REPLICA IDENTITY FULL;CREATE PUBLICATION "%s" FOR TABLE %s`, quotedTableName, fullSlotName, quotedTableName)

	if event != nil {
		operationName, err := d.getOperationNameForEvent(event)
		if err != nil {
			return "", err
		}
		rawSql += fmt.Sprintf(` WITH (publish = '%s')`, operationName)
	}
	return rawSql + ";", nil
}

func (d *Driver) getAlterPublicationEventsSql(publication *activePublication) (string, error) {
	if publication == nil {
		return "", errors.New("publication is nil")
	}

	var events []string
	for targetEvent := types.Operation(1); targetEvent != 0 && targetEvent <= types.OperationAll; targetEvent <<= 1 {
		if (*publication.events)&targetEvent == 0 {
			continue
		}
		operation, err := d.getOperationNameForEvent(&targetEvent)
		if err != nil {
			return "", err
		}
		events = append(events, operation)
	}

	return fmt.Sprintf(`ALTER PUBLICATION "%s" SET (publish = '%s');`, publication.slotName, strings.Join(events, ", ")), nil
}

func (d *Driver) getDropPublicationSlotSql(fullSlotName string) string {
	return fmt.Sprintf(`DROP PUBLICATION IF EXISTS "%s";`, fullSlotName)
}
func (d *Driver) getOperationNameForEvent(e *types.Operation) (string, error) {
	operation := ""
	switch *e {
	case types.OperationInsert:
		operation = "insert"
	case types.OperationUpdate:
		operation = "update"
	case types.OperationDelete:
		operation = "delete"
	case types.OperationTruncate:
		operation = "truncate"
	default:
		return "", errors.New("could not determine event type")
	}
	return operation, nil
}

// Returns tablename as format public.posts.
// posts -> public.posts
// "stats"."name" -> stats.name
// public."posts" -> public.posts
func (d *Driver) sanitizeTableName(tableName string, quote bool) string {
	splits := strings.Split(tableName, ".")
	if len(splits) == 1 {
		splits = []string{"public", strings.ReplaceAll(splits[0], `"`, "")}
	} else {
		splits = []string{strings.ReplaceAll(splits[0], `"`, ""), strings.ReplaceAll(splits[1], `"`, "")}
	}

	if quote {
		splits[0] = `"` + splits[0] + `"`
		splits[1] = `"` + splits[1] + `"`
	}
	return strings.Join(splits, ".")
}

func (d *Driver) sqlExec(conn *pgconn.PgConn, query string) ([]*pgconn.Result, error) {
	d._clientConfig.Logger.Trace().Str("query", query).Msg("sending sql request")
	result := conn.Exec(context.Background(), query)
	return result.ReadAll()
}
