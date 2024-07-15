package trigger

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/quix-labs/flash"
	"strings"
	"time"
)

func (d *Driver) getCreateTriggerSqlForOperation(listenerUid string, l *flash.ListenerConfig, e *flash.Operation) (string, string, error) {
	uniqueName, err := d.getUniqueIdentifierForListenerEvent(listenerUid, e)
	if err != nil {
		return "", "", err
	}

	operation, err := e.StrictName()
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
				PERFORM pg_notify('%s', JSONB_BUILD_OBJECT('old',to_jsonb(OLD),'new',to_jsonb(NEW))::TEXT);
				RETURN COALESCE(NEW, OLD);
			END;
			$trigger$ LANGUAGE plpgsql VOLATILE;`,
			d.Config.Schema, triggerFnName, eventName)
	} else {
		var rawFields, rawConditionSql string

		switch operation {
		case "TRUNCATE":
			rawFields = "null"
		case "DELETE":

			if len(l.Conditions) > 0 {
				rawConditionSql, err = d.getConditionsSql(l.Conditions, "OLD")
				if err != nil {
					return "", "", err
				}
			}

			jsonFields := make([]string, len(l.Fields))
			for i, field := range l.Fields {
				jsonFields[i] = fmt.Sprintf(`'%s', OLD."%s"`, field, field)
			}
			rawFields = fmt.Sprintf(`JSONB_BUILD_OBJECT('old',JSONB_BUILD_OBJECT(%s))::TEXT`, strings.Join(jsonFields, ","))
		case "INSERT":

			if len(l.Conditions) > 0 {
				rawConditionSql, err = d.getConditionsSql(l.Conditions, "NEW")
				if err != nil {
					return "", "", err
				}
			}

			jsonFields := make([]string, len(l.Fields))
			for i, field := range l.Fields {
				jsonFields[i] = fmt.Sprintf(`'%s', NEW."%s"`, field, field)
			}
			rawFields = fmt.Sprintf(`JSONB_BUILD_OBJECT('new',JSONB_BUILD_OBJECT(%s))::TEXT`, strings.Join(jsonFields, ","))
		case "UPDATE":
			oldJsonFields := make([]string, len(l.Fields))
			newJsonFields := make([]string, len(l.Fields))
			for i, field := range l.Fields {
				oldJsonFields[i] = fmt.Sprintf(`'%s', OLD."%s"`, field, field)
				newJsonFields[i] = fmt.Sprintf(`'%s', NEW."%s"`, field, field)
			}

			// Build raw conditions for field updates
			rawConditions := make([]string, len(l.Fields))
			for i, field := range l.Fields {
				rawConditions[i] = fmt.Sprintf(`(OLD."%s" IS DISTINCT FROM NEW."%s")`, field, field)
			}
			rawConditionSql = strings.Join(rawConditions, " OR ")

			// Build conditions for soft delete check
			var oldConditionsSql, newConditionsSql string = "null", "null"
			if len(l.Conditions) > 0 {
				oldConditionsSql, err = d.getConditionsSql(l.Conditions, "OLD")
				if err != nil {
					return "", "", err
				}
				newConditionsSql, err = d.getConditionsSql(l.Conditions, "NEW")
				if err != nil {
					return "", "", err
				}

				// Combine update conditions with soft delete conditions
				rawConditionSql = fmt.Sprintf(`((%s)!=(%s)) OR (%s)`, oldConditionsSql, newConditionsSql, rawConditionSql)
			}

			rawFields = fmt.Sprintf(
				`JSONB_BUILD_OBJECT('old',JSONB_BUILD_OBJECT(%s),'new',JSONB_BUILD_OBJECT(%s),'old_condition',%s,'new_condition',%s)::TEXT`,
				strings.Join(oldJsonFields, ","),
				strings.Join(newJsonFields, ","),
				oldConditionsSql,
				newConditionsSql,
			)
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
			DROP TRIGGER IF EXISTS "%s" ON %s;
			CREATE TRIGGER "%s" AFTER %s ON %s FOR EACH ROW EXECUTE PROCEDURE "%s"."%s"();`,
			triggerName, d.sanitizeTableName(l.Table), triggerName, operation, d.sanitizeTableName(l.Table), d.Config.Schema, triggerFnName)
	} else {
		statement += fmt.Sprintf(`
			DROP TRIGGER IF EXISTS "%s" ON %s;
			CREATE TRIGGER "%s" BEFORE TRUNCATE ON %s FOR EACH STATEMENT EXECUTE PROCEDURE "%s"."%s"();`,
			triggerName, d.sanitizeTableName(l.Table), triggerName, d.sanitizeTableName(l.Table), d.Config.Schema, triggerFnName)
	}

	return statement, eventName, nil
}

func (d *Driver) getDeleteTriggerSqlForEvent(listenerUid string, l *flash.ListenerConfig, e *flash.Operation) (string, string, error) {
	uniqueName, err := d.getUniqueIdentifierForListenerEvent(listenerUid, e)
	if err != nil {
		return "", "", err
	}

	triggerFnName := uniqueName + "_fn"
	eventName := uniqueName + "_event"

	return fmt.Sprintf(`DROP FUNCTION IF EXISTS "%s"."%s" CASCADE;`, d.Config.Schema, triggerFnName), eventName, nil
}

func (d *Driver) getUniqueIdentifierForListenerEvent(listenerUid string, e *flash.Operation) (string, error) {
	operationName, err := e.StrictName()
	if err != nil {
		return "", err
	}
	return strings.Join([]string{
		d.Config.Schema,
		listenerUid,
		strings.ToLower(operationName),
	}, "_"), nil
}
func (d *Driver) parseEventName(channel string) (string, flash.Operation, error) {
	parts := strings.Split(channel, "_")
	if len(parts) != 4 {
		return "", 0, errors.New("could not determine unique identifier")
	}

	listenerUid := parts[1]
	operation, err := flash.OperationFromName(parts[2])
	if err != nil {
		return "", 0, err
	}

	return listenerUid, operation, nil

}
func (d *Driver) sanitizeTableName(tableName string) string {
	segments := strings.Split(tableName, ".")
	for i, segment := range segments {
		segments[i] = `"` + segment + `"`
	}
	return strings.Join(segments, ".")
}
func (d *Driver) sqlExec(conn *sql.DB, query string) (sql.Result, error) {
	d._clientConfig.Logger.Trace().Str("query", query).Msg("sending sql request")
	return conn.Exec(query)
}

func (d *Driver) getConditionsSql(conditions []*flash.ListenerCondition, table string) (string, error) {
	rawConditions := make([]string, len(conditions))

	for i, condition := range conditions {
		operator := " IS "
		valueRepr := ""
		// TODO MULTI OPERATOR

		switch condition.Value.(type) {
		case nil:
			valueRepr = "NULL"
		case bool:
			if condition.Value.(bool) == true {
				valueRepr = "TRUE"
			} else {
				valueRepr = "FALSE"
			}
		case string, time.Time:
			valueRepr = fmt.Sprintf(`'%s'`, condition.Value)
		case float32, float64:
			valueRepr = fmt.Sprintf(`%f`, condition.Value)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			valueRepr = fmt.Sprintf(`%d`, condition.Value)
		default:
			return "", errors.New("could not convert condition value to sql")
		}

		rawConditions[i] = fmt.Sprintf(`%s."%s"%s%s`, table, condition.Column, operator, valueRepr)

	}
	return strings.Join(rawConditions, " AND "), nil
}
