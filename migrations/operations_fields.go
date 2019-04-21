package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
)

type AddFields struct {
	Model  string
	Fields gomodels.Fields
}

func (op AddFields) Name() string {
	return "AddFields"
}

func (op AddFields) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op AddFields) SetState(state *AppState) error {
	if _, ok := state.Models[op.Model]; !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	fields := state.Models[op.Model].Fields()
	for name, field := range op.Fields {
		if _, found := fields[name]; found {
			return fmt.Errorf("%s: duplicate field: %s", op.Model, name)
		}
		fields[name] = field
	}
	delete(state.Models, op.Model)
	state.Models[op.Model] = gomodels.New(op.Model, fields)
	return nil
}

type RemoveFields struct {
	Model  string
	Fields []string
}

func (op RemoveFields) Name() string {
	return "RemoveFields"
}

func (op RemoveFields) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op RemoveFields) SetState(state *AppState) error {
	if _, ok := state.Models[op.Model]; !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	fields := state.Models[op.Model].Fields()
	for _, name := range op.Fields {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("%s: field not found: %s", op.Model, name)
		}
		delete(fields, name)
	}
	delete(state.Models, op.Model)
	state.Models[op.Model] = gomodels.New(op.Model, fields)
	return nil
}
