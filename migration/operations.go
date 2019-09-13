package migration

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"reflect"
)

// The Operation interface represents a change in the application state (models,
// fields, indexes...) and how to propagate it to the database schema.
//
// Custom operations can be registered using the RegisterOperation function.
type Operation interface {
	// OpName returns the operation name, which must be unique.
	OpName() string
	// SetState applies the operation changes to the given application state.
	SetState(state *AppState) error
	// Run applies the operation changes to the database schema. The method
	// is automatically called when a Node is migrating forwards.
	//
	// The engine is the database (or transaction if supported) Engine.
	//
	// The state is the application state resulted from the operation changes.
	//
	// The prevState is the application state previous to the operation changes.
	Run(engine gomodel.Engine, state *AppState, prevState *AppState) error
	// Backwards reverse the operation changes on the database schema. The
	// method is automatically called when a Node is migrating backwards.
	//
	// The engine is the database (or transaction if supported) Engine.
	//
	// The state is the application state resulted from the operation changes.
	//
	// The prevState is the application state previous to the operation changes.
	Backwards(engine gomodel.Engine, state *AppState, prevState *AppState) error
}

// OperationList represents the list of operations of a migration Node.
type OperationList []Operation

// MarshalJSON implements the json.Marshaler interface for the OperationList
// type. The type is serialized into a list of JSON objects, where the key is
// the name of the Operation returned by the OpName method and the value is
// the serliazed type implementing the Operation interface.
func (opList OperationList) MarshalJSON() ([]byte, error) {
	result := []map[string]Operation{}
	for _, op := range opList {
		m := map[string]Operation{}
		m[op.OpName()] = op
		result = append(result, m)
	}
	return json.Marshal(result)
}

// UnmarshalJSON implements the json.Unmarshaler interface. An error is
// returned if the name of the operation is not registered.
func (opList *OperationList) UnmarshalJSON(data []byte) error {
	ops := *opList
	rawList := []map[string]json.RawMessage{}
	err := json.Unmarshal(data, &rawList)
	if err != nil {
		return err
	}
	for _, rawMap := range rawList {
		for name, rawOp := range rawMap {
			operation, ok := operationsRegistry[name]
			if !ok {
				return fmt.Errorf("invalid operation: %s", name)
			}
			// Gets the operation type and creates a new value.
			ot := reflect.Indirect(reflect.ValueOf(operation)).Type()
			op := reflect.New(ot).Interface()
			if err := json.Unmarshal(rawOp, op); err != nil {
				return err
			}
			ops = append(ops, op.(Operation))
		}
	}
	*opList = ops
	return nil
}

// operationsRegistry holds a global registry of available operations.
var operationsRegistry = map[string]Operation{
	"CreateModel":  CreateModel{},
	"DeleteModel":  DeleteModel{},
	"AddFields":    AddFields{},
	"RemoveFields": RemoveFields{},
	"AddIndex":     AddIndex{},
	"RemoveIndex":  RemoveIndex{},
}

// RegisterOperation registers a custom operation. Returns an error if the
// operation name returned by the OpName method already exists.
func RegisterOperation(op Operation) error {
	name := op.OpName()
	if _, found := operationsRegistry[name]; found {
		return fmt.Errorf("migrations: duplicate operation: %s", name)
	}
	operationsRegistry[name] = op
	return nil
}
