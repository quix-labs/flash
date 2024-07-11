package wal_logical

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/quix-labs/flash"
	"strings"
)

func (d *Driver) getFullSlotName(slotName string) string {
	return d.Config.PublicationSlotPrefix + "-" + slotName
}

func (d *Driver) getCreatePublicationSlotSql(fullSlotName string, config *flash.ListenerConfig, operation *flash.Operation) (string, error) {
	if config == nil {
		return fmt.Sprintf(`CREATE PUBLICATION "%s";`, fullSlotName), nil
	}

	rawSql := d.getDropPublicationSlotSql(fullSlotName)
	// SET REPLICA IDENTITY TO FULL ON CREATION
	quotedTableName := d.sanitizeTableName(config.Table, true)
	rawSql += fmt.Sprintf(`ALTER TABLE %s REPLICA IDENTITY FULL;CREATE PUBLICATION "%s" FOR TABLE %s`, quotedTableName, fullSlotName, quotedTableName)

	if operation != nil {
		//TODO THROW ERROR IF NOT ATOMIC OR JOIN EACH ATOMIC (see .getAlterPublicationEventsSql() )
		operationName, err := operation.StrictName()
		if err != nil {
			return "", err
		}
		rawSql += fmt.Sprintf(` WITH (publish = '%s')`, strings.ToLower(operationName))
	}
	return rawSql + ";", nil
}

func (d *Driver) getAlterPublicationEventsSql(publication *activePublication) (string, error) {
	if publication == nil {
		return "", errors.New("publication is nil")
	}

	var rawOperations []string
	for _, targetOperation := range publication.operations.GetAtomics() {
		operation, err := targetOperation.StrictName()
		if err != nil {
			return "", err
		}
		rawOperations = append(rawOperations, strings.ToLower(operation))
	}

	return fmt.Sprintf(`ALTER PUBLICATION "%s" SET (publish = '%s');`, publication.slotName, strings.Join(rawOperations, ", ")), nil
}

func (d *Driver) getDropPublicationSlotSql(fullSlotName string) string {
	return fmt.Sprintf(`DROP PUBLICATION IF EXISTS "%s";`, fullSlotName)
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
