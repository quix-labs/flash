package flash

type EventData map[string]any
type Event interface {
	GetOperation() Operation
}

type InsertEvent struct {
	New *EventData
}
type UpdateEvent struct {
	Old *EventData
	New *EventData
}
type DeleteEvent struct {
	Old *EventData
}
type TruncateEvent struct{}

func (e *InsertEvent) GetOperation() Operation {
	return OperationInsert
}
func (e *UpdateEvent) GetOperation() Operation {
	return OperationUpdate
}
func (e *DeleteEvent) GetOperation() Operation {
	return OperationDelete
}
func (e *TruncateEvent) GetOperation() Operation {
	return OperationTruncate
}
