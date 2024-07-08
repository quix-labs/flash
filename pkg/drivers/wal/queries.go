package wal

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

func (d *Driver) getCreatePublicationSlotSql(fullSlotName string, config *types.ListenerConfig, event *types.Event) (string, error) {
	rawSql := fmt.Sprintf(`CREATE PUBLICATION "%s"`, fullSlotName)
	if config != nil {
		rawSql += fmt.Sprintf(` FOR TABLE %s`, d.sanitizeTableName(config.Table, true))
		//TODO ADD ALTER TABLE x REPLICA IDENTITY FULL
	}
	if event != nil {
		operationName, err := d.getOperationNameForEvent(event)
		if err != nil {
			return "", err
		}
		rawSql += fmt.Sprintf(` WITH (publish = '%s')`, operationName)
	}
	return rawSql + ";", nil
}

func (d *Driver) getDropPublicationSlotSql(fullSlotName string) string {
	return fmt.Sprintf(`DROP PUBLICATION IF EXISTS "%s";`, fullSlotName)
}
func (d *Driver) getOperationNameForEvent(e *types.Event) (string, error) {
	operation := ""
	switch *e {
	case types.EventInsert:
		operation = "insert"
	case types.EventUpdate:
		operation = "update"
	case types.EventDelete:
		operation = "delete"
	case types.EventTruncate:
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
