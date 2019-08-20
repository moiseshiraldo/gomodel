package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"reflect"
)

type Operation interface {
	OpName() string
	SetState(state *AppState) error
	Run(tx *gomodels.Transaction, state *AppState, prevState *AppState) error
	Backwards(tx *gomodels.Transaction, state *AppState, prevState *AppState) error
}

type OperationList []Operation

func (opList OperationList) MarshalJSON() ([]byte, error) {
	result := []map[string]Operation{}
	for _, op := range opList {
		m := map[string]Operation{}
		m[op.OpName()] = op
		result = append(result, m)
	}
	return json.Marshal(result)
}

func (op *OperationList) UnmarshalJSON(data []byte) error {
	opList := *op
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
			if err := json.Unmarshal(rawOp, operation); err != nil {
				return err
			}
			opVal := reflect.ValueOf(operation).Interface()
			if op, ok := opVal.(Operation); ok {
				opList = append(opList, op)
			} else {
				op := reflect.Indirect(reflect.ValueOf(operation)).Interface()
				opList = append(opList, op.(Operation))
			}
		}
	}
	*op = opList
	return nil
}

var operationsRegistry = map[string]Operation{
	"CreateModel":  &CreateModel{},
	"DeleteModel":  &DeleteModel{},
	"AddFields":    &AddFields{},
	"RemoveFields": &RemoveFields{},
	"AddIndex":     &AddIndex{},
	"RemoveIndex":  &RemoveIndex{},
}

func RegisterOperation(name string, op Operation) error {
	if _, found := operationsRegistry[name]; found {
		return fmt.Errorf("migrations: duplicate operation: %s", name)
	}
	operationsRegistry[name] = op
	return nil
}
