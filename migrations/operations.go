package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
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

type CreateModel struct {
	Model  string
	Fields gomodels.Fields
}

func (op CreateModel) Name() string {
	return "CreateModel"
}

func (op CreateModel) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op CreateModel) SetState(state *AppState) error {
	if _, found := state.Models[op.Model]; found {
		return fmt.Errorf("duplicate model: %s", op.Model)
	}
	state.Models[op.Model] = gomodels.New(op.Model, op.Fields)
	return nil
}

type Field struct {
	Type    string
	Options gomodels.Field
}

func (f *Field) UnmarshalJSON(data []byte) error {
	obj := struct {
		Type    string
		Options json.RawMessage
	}{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	f.Type = obj.Type
	native, ok := gomodels.AvailableFields()[obj.Type]
	if !ok {
		return fmt.Errorf("invalid field type: %s", obj.Type)
	}
	f.Options, err = native.FromJSON(obj.Options)
	if err != nil {
		return err
	}
	return nil
}

func AvailableOperations() map[string]Operation {
	return map[string]Operation{
		"CreateModel": CreateModel{},
	}
}
