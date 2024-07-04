package client

import (
	"errors"
	"fmt"
)

/** ----------------------------- SQL UTILS - CAN BE EXTERNALIZED -------------------------- */

// GetCreateTriggerSqlForEvent generate sql query for event trigger, can be external
func GetCreateTriggerSqlForEvent(l *Listener, e Event, uniqueName string, schema string) (string, error) {
	operation := ""
	switch e {
	case EventInsert:
		operation = "INSERT"
	case EventUpdate:
		operation = "UPDATE"
	case EventDelete:
		operation = "DELETE"
	case EventTruncate:
		operation = "TRUNCATE"
	default:
		return "", errors.New("could not determine event type")
	}

	triggerFnName := fmt.Sprintf("%s_fn", uniqueName)

	statement := fmt.Sprintf(`
CREATE OR REPLACE FUNCTION "%s"."%s"() RETURNS trigger AS $trigger$
BEGIN 
  PERFORM pg_notify('%s', ROW_TO_JSON(COALESCE(NEW,OLD)));
  RETURN COALESCE(NEW, OLD);
END;
$trigger$ LANGUAGE plpgsql VOLATILE;`,
		schema, triggerFnName, uniqueName)
	statement += fmt.Sprintf(
		`CREATE OR REPLACE TRIGGER "%s" BEFORE %s ON %s FOR EACH ROW EXECUTE PROCEDURE "%s"."%s"();`,
		uniqueName,
		operation,
		l.Config.Table,
		schema,
		triggerFnName,
	)
	return statement, nil
}

// GetDeleteTriggerSqlForEvent generate sql query for event trigger, can be external
func GetDeleteTriggerSqlForEvent(l *Listener, e Event, uniqueName string, schema string) (string, error) {
	triggerFnName := fmt.Sprintf("%s_fn", uniqueName)
	return fmt.Sprintf(`DROP FUNCTION IF EXISTS "%s"."%s" CASCADE;`, schema, triggerFnName), nil
}
