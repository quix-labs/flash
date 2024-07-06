package trigger

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/quix-labs/flash/pkg/types"
	"strings"
)

func (d *Driver) getCreateTriggerSqlForEvent(listenerUid string, l *types.ListenerConfig, e *types.Event) (string, string, error) {
	uniqueName, err := d.getUniqueIdentifierForListenerEvent(listenerUid, e)
	if err != nil {
		return "", "", err
	}

	operation, err := d.getOperationNameForEvent(e)
	if err != nil {
		return "", "", err
	}

	triggerName := uniqueName + "_trigger"
	triggerFnName := uniqueName + "_fn"
	eventName := uniqueName + "_event"

	var statement string
	if len(l.Fields) == 0 {
		statement = fmt.Sprintf(`
			CREATE OR REPLACE FUNCTION "%s"."%s"() RETURNS trigger AS $trigger$
			BEGIN 
				PERFORM pg_notify('%s', ROW_TO_JSON(COALESCE(NEW, OLD))::TEXT);
				RETURN COALESCE(NEW, OLD);
			END;
			$trigger$ LANGUAGE plpgsql VOLATILE;`,
			d.Config.Schema, triggerFnName, eventName)
	} else {
		var rawFields, rawConditionSql string

		if operation == "TRUNCATE" {
			rawFields = "null"
		} else {
			rawConditions := make([]string, len(l.Fields))
			for i, field := range l.Fields {
				rawConditions[i] = fmt.Sprintf(`OLD."%s" <> NEW."%s"`, field, field)
			}
			rawConditionSql = strings.Join(rawConditions, " OR ")

			jsonFields := make([]string, len(l.Fields))
			for i, field := range l.Fields {
				jsonFields[i] = fmt.Sprintf(`'%s', COALESCE(NEW."%s", OLD."%s")`, field, field, field)
			}
			rawFields = fmt.Sprintf(`JSON_BUILD_OBJECT(%s)::TEXT`, strings.Join(jsonFields, ","))
		}

		if rawConditionSql == "" {
			statement = fmt.Sprintf(`
				CREATE OR REPLACE FUNCTION "%s"."%s"() RETURNS trigger AS $trigger$
				BEGIN 
					PERFORM pg_notify('%s', %s);
					RETURN COALESCE(NEW, OLD);
				END;
				$trigger$ LANGUAGE plpgsql VOLATILE;`,
				d.Config.Schema, triggerFnName, eventName, rawFields)
		} else {
			statement = fmt.Sprintf(`
				CREATE OR REPLACE FUNCTION "%s"."%s"() RETURNS trigger AS $trigger$
				BEGIN
					IF %s THEN
						PERFORM pg_notify('%s', %s);
					END IF;
					RETURN COALESCE(NEW, OLD);
				END;
				$trigger$ LANGUAGE plpgsql VOLATILE;`,
				d.Config.Schema, triggerFnName, rawConditionSql, eventName, rawFields)
		}
	}

	if operation != "TRUNCATE" {
		statement += fmt.Sprintf(`
			CREATE OR REPLACE TRIGGER "%s" BEFORE %s ON %s FOR EACH ROW EXECUTE PROCEDURE "%s"."%s"();`,
			triggerName, operation, d.sanitizeTableName(l.Table), d.Config.Schema, triggerFnName)
	} else {
		statement += fmt.Sprintf(`
			CREATE OR REPLACE TRIGGER "%s" BEFORE TRUNCATE ON %s FOR EACH STATEMENT EXECUTE PROCEDURE "%s"."%s"();`,
			triggerName, d.sanitizeTableName(l.Table), d.Config.Schema, triggerFnName)
	}

	return statement, eventName, nil
}

func (d *Driver) getDeleteTriggerSqlForEvent(listenerUid string, l *types.ListenerConfig, e *types.Event) (string, string, error) {
	uniqueName, err := d.getUniqueIdentifierForListenerEvent(listenerUid, e)
	if err != nil {
		return "", "", err
	}

	triggerFnName := uniqueName + "_fn"
	eventName := uniqueName + "_event"

	return fmt.Sprintf(`DROP FUNCTION IF EXISTS "%s"."%s" CASCADE;`, d.Config.Schema, triggerFnName), eventName, nil
}

func (d *Driver) getOperationNameForEvent(e *types.Event) (string, error) {
	operation := ""
	switch *e {
	case types.EventInsert:
		operation = "INSERT"
	case types.EventUpdate:
		operation = "UPDATE"
	case types.EventDelete:
		operation = "DELETE"
	case types.EventTruncate:
		operation = "TRUNCATE"
	default:
		return "", errors.New("could not determine event type")
	}
	return operation, nil
}
func (d *Driver) getEventForOperationName(operationName string) (types.Event, error) {
	var event types.Event

	switch strings.ToUpper(operationName) {
	case "INSERT":
		event = types.EventInsert
	case "UPDATE":
		event = types.EventUpdate
	case "DELETE":
		event = types.EventDelete
	case "TRUNCATE":
		event = types.EventTruncate
	default:
		return 0, errors.New("could not determine event type")
	}
	return event, nil
}
func (d *Driver) getUniqueIdentifierForListenerEvent(listenerUid string, e *types.Event) (string, error) {
	operationName, err := d.getOperationNameForEvent(e)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{
		d.Config.Schema,
		listenerUid,
		strings.ToLower(operationName),
	}, "_"), nil
}
func (d *Driver) parseEventName(channel string) (string, types.Event, error) {
	parts := strings.Split(channel, "_")
	if len(parts) != 4 {
		return "", 0, errors.New("could not determine unique identifier")
	}

	listenerUid := parts[1]
	event, err := d.getEventForOperationName(parts[2])
	if err != nil {
		return "", 0, err
	}

	return listenerUid, event, nil

}
func (d *Driver) sanitizeTableName(tableName string) string {
	segments := strings.Split(tableName, ".")
	for i, segment := range segments {
		segments[i] = `"` + segment + `"`
	}
	return strings.Join(segments, ".")
}
func (d *Driver) sqlExec(conn *pgx.Conn, query string) (pgconn.CommandTag, error) {
	d._clientConfig.Logger.Trace().Str("query", query).Msg("sending sql request")
	return conn.Exec(context.TODO(), query)
}
