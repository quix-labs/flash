package flash

import (
	"errors"
	"strings"
)

type Operation uint8

const (
	OperationInsert Operation = 1 << iota
	OperationUpdate
	OperationDelete
	OperationTruncate
)
const (
	OperationAll = OperationInsert | OperationUpdate | OperationDelete | OperationTruncate
)

func (o Operation) IsAtomic() bool {
	return o == OperationInsert ||
		o == OperationUpdate ||
		o == OperationDelete ||
		o == OperationTruncate
}

func (o Operation) GetAtomics() []Operation {
	var operations []Operation
	for mask := OperationInsert; mask != 0 && mask <= OperationTruncate; mask <<= 1 {
		if o&mask != 0 {
			operations = append(operations, mask)
		}
	}
	return operations
}

// IncludeAll checks if the current operation includes all specified atomic operations.
func (o Operation) IncludeAll(targetOperation Operation) bool {
	return o&targetOperation == targetOperation
}

// IncludeOne checks if the current operation includes at least one of the specified atomic operations.
func (o Operation) IncludeOne(targetOperation Operation) bool {
	return o&targetOperation > 0
}

// StrictName returns the name of the operation, or throws an error if it doesn't exist
func (o Operation) StrictName() (string, error) {
	switch o {
	case OperationInsert:
		return "INSERT", nil
	case OperationUpdate:
		return "UPDATE", nil
	case OperationDelete:
		return "DELETE", nil
	case OperationTruncate:
		return "TRUNCATE", nil
	default:
		return "UNKNOWN", errors.New("unknown operation")
	}
}

// Use with caution, because no errors are returned when invalid
func (o Operation) String() string {
	if o.IsAtomic() {
		name, _ := o.StrictName()
		return name
	} else {
		atomicString := []string{}
		for _, atomicOperation := range o.GetAtomics() {
			atomicString = append(atomicString, atomicOperation.String())
		}
		if len(atomicString) > 1 {
			return strings.Join(atomicString, " | ")
		} else {
			return "UNKNOWN"
		}
	}
}

func OperationFromName(name string) (Operation, error) {
	switch strings.ToUpper(name) {
	case "INSERT":
		return OperationInsert, nil
	case "UPDATE":
		return OperationUpdate, nil
	case "DELETE":
		return OperationDelete, nil
	case "TRUNCATE":
		return OperationTruncate, nil
	default:
		return 0, errors.New("unknown operation name")
	}
}
