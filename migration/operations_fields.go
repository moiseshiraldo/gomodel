package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
)

type AddFields struct {
	Model  string
	Fields gomodel.Fields
}

func (op AddFields) OpName() string {
	return "AddFields"
}

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

func (op AddFields) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.AddColumns(state.Models[op.Model], op.Fields)
}

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

type RemoveFields struct {
	Model  string
	Fields []string
}

func (op RemoveFields) OpName() string {
	return "RemoveFields"
}

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

func (op RemoveFields) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropColumns(prevState.Models[op.Model], op.Fields...)
}

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
	return engine.AddColumns(state.Models[op.Model], newFields)
}
