package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
)

type CreateModel struct {
	Name   string
	Table  string `json:",omitempty"`
	Fields gomodels.Fields
}

func (op CreateModel) OpName() string {
	return "CreateModel"
}

func (op *CreateModel) SetState(state *AppState) error {
	if _, found := state.models[op.Name]; found {
		return fmt.Errorf("duplicate model: %s", op.Name)
	}
	table := op.Table
	if table == "" {
		table = fmt.Sprintf("%s_%s", state.app.Name(), op.Name)
	}
	state.models[op.Name] = gomodels.New(
		op.Name, op.Fields, gomodels.Options{Table: table},
	).Model
	return nil
}

func (op CreateModel) Run(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	return tx.CreateTable(state.models[op.Name])
}

func (op CreateModel) Backwards(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	return tx.DropTable(state.models[op.Name])
}

type DeleteModel struct {
	Name string
}

func (op DeleteModel) OpName() string {
	return "DeleteModel"
}

func (op *DeleteModel) SetState(state *AppState) error {
	if _, ok := state.models[op.Name]; !ok {
		return fmt.Errorf("model not found: %s", op.Name)
	}
	delete(state.models, op.Name)
	return nil
}

func (op DeleteModel) Run(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	return tx.DropTable(prevState.models[op.Name])
}

func (op DeleteModel) Backwards(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	return tx.CreateTable(prevState.models[op.Name])
}

type AddIndex struct {
	Model  string
	Name   string
	Fields []string
}

func (op AddIndex) OpName() string {
	return "AddIndex"
}

func (op *AddIndex) SetState(state *AppState) error {
	model, ok := state.models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	indexes := model.Indexes()
	if _, found := model.Indexes()[op.Name]; found {
		return fmt.Errorf("duplicate index name: %s", op.Name)
	}
	indexes[op.Name] = op.Fields
	options := gomodels.Options{Table: model.Table(), Indexes: indexes}
	state.models[op.Model] = gomodels.New(
		model.Name(), model.Fields(), options,
	).Model
	return nil
}

func (op AddIndex) Run(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	return tx.AddIndex(state.models[op.Model], op.Name, op.Fields...)
}

func (op AddIndex) Backwards(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	return tx.DropIndex(state.models[op.Model], op.Name)
}

type RemoveIndex struct {
	Model string
	Name  string
}

func (op RemoveIndex) OpName() string {
	return "RemoveIndex"
}

func (op *RemoveIndex) SetState(state *AppState) error {
	model, ok := state.models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	indexes := model.Indexes()
	if _, ok := model.Indexes()[op.Name]; !ok {
		return fmt.Errorf("index not found: %s", op.Name)
	}
	delete(indexes, op.Name)
	options := gomodels.Options{Table: model.Table(), Indexes: indexes}
	state.models[op.Model] = gomodels.New(
		model.Name(), model.Fields(), options,
	).Model
	return nil
}

func (op RemoveIndex) Run(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	return tx.DropIndex(state.models[op.Model], op.Name)
}

func (op RemoveIndex) Backwards(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	model := prevState.models[op.Model]
	return tx.AddIndex(model, op.Name, model.Indexes()[op.Name]...)
}
