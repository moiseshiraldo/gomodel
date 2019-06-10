package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type Operation interface {
	OpName() string
	SetState(state *AppState) error
	Run(tx *sql.Tx, app string, driver string) error
	Backwards(tx *sql.Tx, app string, driver string, prevState *AppState) error
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
			operation, ok := AvailableOperations()[name]
			if !ok {
				return fmt.Errorf("invalid operation: %s", name)
			}
			if err := json.Unmarshal(rawOp, &operation); err != nil {
				return err
			}
			opList = append(opList, operation)
		}
	}
	*op = opList
	return nil
}

func AvailableOperations() map[string]Operation {
	return map[string]Operation{
		"CreateModel":  &CreateModel{},
		"DeleteModel":  &DeleteModel{},
		"AddFields":    &AddFields{},
		"RemoveFields": &RemoveFields{},
		"AddIndex":     &AddIndex{},
		"RemoveIndex":  &RemoveIndex{},
	}
}
