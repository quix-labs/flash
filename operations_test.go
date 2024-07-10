package flash

import "testing"

func TestIsAtomic(t *testing.T) {
	tests := []struct {
		name     string
		o        Operation
		expected bool
	}{
		{"Atomic Operation", OperationTruncate, true},
		{"Composite Operation", OperationInsert | OperationUpdate, false},
		{"Atomic But Invalid", 32, false},
		{"Empty Operation", 0, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.o.IsAtomic() != test.expected {
				t.Errorf("IsAtomic() failed for %v: expected %v, got %v", test.o, test.expected, test.o.IsAtomic())
			}
		})
	}
}

func TestGetAtomics(t *testing.T) {
	tests := []struct {
		name     string
		o        Operation
		expected []Operation
	}{
		{"Atomic Operation", OperationTruncate, []Operation{OperationTruncate}},
		{"Composite Operation", OperationInsert | OperationUpdate, []Operation{OperationInsert, OperationUpdate}},
		{"Composite All Operation", OperationAll, []Operation{OperationInsert, OperationUpdate, OperationDelete, OperationTruncate}},
		{"Empty Operation", 0, []Operation{}},
		{"Unknown Atomic", 32, []Operation{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			atomics := test.o.GetAtomics()
			if len(atomics) != len(test.expected) {
				t.Errorf("GetAtomics() failed for %v: expected length %v, got length %v", test.o, len(test.expected), len(atomics))
			} else {
				for i, op := range atomics {
					if op != test.expected[i] {
						t.Errorf("GetAtomics() failed for %v: expected %v at index %d, got %v", test.o, test.expected[i], i, op)
					}
				}
			}
		})
	}
}

func TestIncludeAll(t *testing.T) {
	tests := []struct {
		name     string
		o        Operation
		mask     Operation
		expected bool
	}{
		{"IncludeAll - true", OperationInsert | OperationUpdate | OperationDelete, OperationInsert | OperationUpdate, true},
		{"IncludeAll - false", OperationInsert | OperationUpdate, OperationInsert | OperationUpdate | OperationDelete, false},
		{"IncludeAll - empty operation", 0, OperationAll, false},
		{"IncludeAll - unknown", 32, OperationAll, false},
		{"IncludeAll - unknown", OperationAll, 32, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.o.IncludeAll(test.mask) != test.expected {
				t.Errorf("IncludeAll() failed for %v with mask %v: expected %v, got %v", test.o, test.mask, test.expected, test.o.IncludeAll(test.mask))
			}
		})
	}
}

func TestIncludeOne(t *testing.T) {
	tests := []struct {
		name     string
		o        Operation
		mask     Operation
		expected bool
	}{
		{"IncludeOne - true", OperationInsert | OperationUpdate | OperationDelete, OperationDelete, true},
		{"IncludeOne - false", OperationInsert | OperationUpdate, OperationTruncate, false},
		{"IncludeOne - empty operation", 0, OperationInsert, false},
		{"IncludeOne - Atomic same", OperationUpdate, OperationUpdate, true},
		{"IncludeOne - Atomic different", OperationUpdate, OperationDelete, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.o.IncludeOne(test.mask) != test.expected {
				t.Errorf("IncludeOne() failed for %v with mask %v: expected %v, got %v", test.o, test.mask, test.expected, test.o.IncludeOne(test.mask))
			}
		})
	}
}

func TestStrictName(t *testing.T) {
	tests := []struct {
		name         string
		o            Operation
		expectedName string
		expectedErr  bool
	}{
		{"Insert Operation", OperationInsert, "INSERT", false},
		{"Update Operation", OperationUpdate, "UPDATE", false},
		{"Delete Operation", OperationDelete, "DELETE", false},
		{"Truncate Operation", OperationTruncate, "TRUNCATE", false},
		{"Unknown Operation", Operation(32), "UNKNOWN", true},
		{"Composite Operation", OperationInsert | OperationUpdate, "UNKNOWN", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, err := test.o.StrictName()

			// Check name correctness
			if name != test.expectedName {
				t.Errorf("StrictName() failed for %v: expected name %v, got %v", test.o, test.expectedName, name)
			}

			// Check error correctness
			if (err != nil) != test.expectedErr {
				t.Errorf("StrictName() failed for %v: expected error %v, got error %v", test.o, test.expectedErr, err)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		o        Operation
		expected string
	}{
		{"Single Atomic Operation", OperationInsert, "INSERT"},
		{"Multiple Atomic Operations", OperationInsert | OperationUpdate | OperationTruncate, "INSERT | UPDATE | TRUNCATE"},
		{"Empty Operation", 0, "UNKNOWN"},
		{"Unknown Operation", 32, "UNKNOWN"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.o.String()

			if result != test.expected {
				t.Errorf("String() failed for %v: expected '%v', got '%v'", test.o, test.expected, result)
			}
		})
	}
}

func TestOperationFromName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Operation
		expectedErr bool
	}{
		{"Insert", "insert", OperationInsert, false},
		{"Update", "update", OperationUpdate, false},
		{"Delete", "delete", OperationDelete, false},
		{"Truncate", "truncate", OperationTruncate, false},
		{"Truncate", "INSERT", OperationInsert, false},
		{"Truncate", "UPDATE", OperationUpdate, false},
		{"Truncate", "DELETE", OperationDelete, false},
		{"Truncate", "TRUNCATE", OperationTruncate, false},
		{"Unknown", "unknown", 0, true},
		{"Empty String", "", 0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := OperationFromName(test.input)

			if (err != nil) != test.expectedErr {
				t.Errorf("OperationFromName() error = %v, expected error = %v", err, test.expectedErr)
				return
			}

			if result != test.expected {
				t.Errorf("OperationFromName() = %v, expected %v", result, test.expected)
			}
		})
	}
}
