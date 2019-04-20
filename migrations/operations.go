package migrations

import (
	"encoding/json"
	"fmt"
)

type Operation interface {
	Name() string
	FromJSON(raw []byte) (Operation, error)
	SetState(state *AppState) error
}

type OperationList []Operation

func (opList OperationList) MarshalJSON() ([]byte, error) {
	result := []map[string]Operation{}
	for _, op := range opList {
		m := map[string]Operation{}
		m[op.Name()] = op
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
			native, ok := AvailableOperations()[name]
			if !ok {
				return fmt.Errorf("invalid operation: %s", name)
			}
			operation, err := native.FromJSON(rawOp)
			if err != nil {
				return err
			}
			opList = append(*op, operation)
		}
	}
	*op = opList
	return nil
}

func AvailableOperations() map[string]Operation {
	return map[string]Operation{
		"CreateModel": CreateModel{},
		"DeleteModel": DeleteModel{},
		"AddFields":   AddFields{},
	}
}
