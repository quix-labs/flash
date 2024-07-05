package trigger

import (
	"errors"
	"fmt"
	"github.com/quix-labs/flash/pkg/types"
	"strings"
)

func (d *Driver) getCreateTriggerSqlForEvent(l *types.ListenerConfig, e *types.Event) (string, error) {
	uniqueName, err := d.getUniqueIdentifierForListenerEvent(l, e)
	if err != nil {
		return "", err
	}

	operation, err := d.getOperationNameForEvent(e)
	if err != nil {
		return "", err
	}

	triggerName := uniqueName + "_trigger"
	triggerFnName := uniqueName + "_fn"

	statement := fmt.Sprintf(`
CREATE OR REPLACE FUNCTION "%s"."%s"() RETURNS trigger AS $trigger$
BEGIN 
  PERFORM pg_notify('%s', ROW_TO_JSON(COALESCE(NEW,OLD)));
  RETURN COALESCE(NEW, OLD);
END;
$trigger$ LANGUAGE plpgsql VOLATILE;`,
		d.Config.Schema, triggerFnName, uniqueName)
	statement += fmt.Sprintf(
		`CREATE OR REPLACE TRIGGER "%s" BEFORE %s ON %s FOR EACH ROW EXECUTE PROCEDURE "%s"."%s"();`,
		triggerName,
		operation,
		l.Table,
		d.Config.Schema,
		triggerFnName,
	)
	return statement, nil
}

func (d *Driver) getDeleteTriggerSqlForEvent(l *types.ListenerConfig, e *types.Event) (string, error) {
	uniqueName, err := d.getUniqueIdentifierForListenerEvent(l, e)
	if err != nil {
		return "", err
	}

	triggerFnName := uniqueName + "_fn"
	return fmt.Sprintf(`DROP FUNCTION IF EXISTS "%s"."%s" CASCADE;`, d.Config.Schema, triggerFnName), nil
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

func (d *Driver) getUniqueIdentifierForListenerEvent(l *types.ListenerConfig, e *types.Event) (string, error) {
	operationName, err := d.getOperationNameForEvent(e)
	if err != nil {
		return "", err
	}
	return d.Config.Schema + "_" + l.Table + "_" + strings.ToLower(operationName), nil
}
