package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
)

// AddFields implements the Operation interface to add new fields.
type AddFields struct {
	Model  string
	Fields gomodel.Fields
}

// OpName returns the operation name.
func (op AddFields) OpName() string {
	return "AddFields"
}

// SetState adds the new fields to the model in the given the application state.
func (op AddFields) SetState(state *AppState) error {
	if _, ok := state.Models[op.Model]; !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	model := state.Models[op.Model]
	for name, field := range op.Fields {
		if err := model.AddField(name, field); err != nil {
			return err
		}
	}
	return nil
}

// Run adds the new columns to the table on the database.
func (op AddFields) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.AddColumns(state.Models[op.Model], op.Fields)
}

// Backwards removes the columns from the table on the database.
func (op AddFields) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	fields := make([]string, 0, len(op.Fields))
	for name := range op.Fields {
		fields = append(fields, name)
	}
	return engine.DropColumns(state.Models[op.Model], fields...)
}

// RemoveFields implements the Operation interface to remove fields.
type RemoveFields struct {
	Model  string
	Fields []string
}

// OpName returns the operation name.
func (op RemoveFields) OpName() string {
	return "RemoveFields"
}

// SetState removes the fields from the model in the given the application state.
func (op RemoveFields) SetState(state *AppState) error {
	if _, ok := state.Models[op.Model]; !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	model := state.Models[op.Model]
	for _, name := range op.Fields {
		if err := model.RemoveField(name); err != nil {
			return err
		}
	}
	return nil
}

// Run removes the columns from the table on the database.
func (op RemoveFields) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropColumns(prevState.Models[op.Model], op.Fields...)
}

// Backwards adds the columns to the table on the database.
func (op RemoveFields) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	fields := prevState.Models[op.Model].Fields()
	newFields := gomodel.Fields{}
	for _, name := range op.Fields {
		newFields[name] = fields[name]
	}
	return engine.AddColumns(prevState.Models[op.Model], newFields)
}
