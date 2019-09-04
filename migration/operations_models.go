package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
)

type CreateModel struct {
	Name   string
	Table  string `json:",omitempty"`
	Fields gomodel.Fields
}

func (op CreateModel) OpName() string {
	return "CreateModel"
}

func (op CreateModel) SetState(state *AppState) error {
	if _, found := state.Models[op.Name]; found {
		return fmt.Errorf("duplicate model: %s", op.Name)
	}
	table := op.Table
	if table == "" {
		table = fmt.Sprintf("%s_%s", state.app.Name(), op.Name)
	}
	state.Models[op.Name] = gomodel.New(
		op.Name, op.Fields, gomodel.Options{Table: table},
	).Model
	return nil
}

func (op CreateModel) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.CreateTable(state.Models[op.Name], true)
}

func (op CreateModel) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropTable(state.Models[op.Name])
}

type DeleteModel struct {
	Name string
}

func (op DeleteModel) OpName() string {
	return "DeleteModel"
}

func (op DeleteModel) SetState(state *AppState) error {
	if _, ok := state.Models[op.Name]; !ok {
		return fmt.Errorf("model not found: %s", op.Name)
	}
	delete(state.Models, op.Name)
	return nil
}

func (op DeleteModel) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropTable(prevState.Models[op.Name])
}

func (op DeleteModel) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.CreateTable(prevState.Models[op.Name], true)
}

type AddIndex struct {
	Model  string
	Name   string
	Fields []string
}

func (op AddIndex) OpName() string {
	return "AddIndex"
}

func (op AddIndex) SetState(state *AppState) error {
	model, ok := state.Models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	return model.AddIndex(op.Name, op.Fields...)
}

func (op AddIndex) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.AddIndex(state.Models[op.Model], op.Name, op.Fields...)
}

func (op AddIndex) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropIndex(state.Models[op.Model], op.Name)
}

type RemoveIndex struct {
	Model string
	Name  string
}

func (op RemoveIndex) OpName() string {
	return "RemoveIndex"
}

func (op RemoveIndex) SetState(state *AppState) error {
	model, ok := state.Models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	return model.RemoveIndex(op.Name)
}

func (op RemoveIndex) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropIndex(state.Models[op.Model], op.Name)
}

func (op RemoveIndex) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	model := prevState.Models[op.Model]
	return engine.AddIndex(model, op.Name, model.Indexes()[op.Name]...)
}
