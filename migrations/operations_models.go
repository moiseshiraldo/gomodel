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
		fmt.Printf(table)
	}
	state.models[op.Name] = gomodels.New(
		op.Name, op.Fields, gomodels.Options{Table: table},
	).Model
	return nil
}

func (op CreateModel) Run(tx *gomodels.Transaction, state *AppState) error {
	if op.Table == "" {
		op.Table = fmt.Sprintf("%s_%s", state.app.Name(), op.Name)
	}
	return tx.CreateTable(op.Table, op.Fields)
}

func (op CreateModel) Backwards(tx *gomodels.Transaction, pS *AppState) error {
	if op.Table == "" {
		op.Table = fmt.Sprintf("%s_%s", pS.app.Name(), op.Name)
	}
	return tx.DropTable(op.Table)
}

type DeleteModel struct {
	Name  string
	table string `json:"-"`
}

func (op DeleteModel) OpName() string {
	return "DeleteModel"
}

func (op *DeleteModel) SetState(state *AppState) error {
	if _, ok := state.models[op.Name]; !ok {
		return fmt.Errorf("model not found: %s", op.Name)
	}
	op.table = state.models[op.Name].Table()
	delete(state.models, op.Name)
	return nil
}

func (op DeleteModel) Run(tx *gomodels.Transaction, state *AppState) error {
	return tx.DropTable(op.table)
}

func (op DeleteModel) Backwards(tx *gomodels.Transaction, pS *AppState) error {
	model := pS.models[op.Name]
	return tx.CreateTable(model.Table(), model.Fields())
}

type AddIndex struct {
	Model  string
	table  string `json:"-"`
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
	op.table = model.Table()
	indexes := model.Indexes()
	if _, found := model.Indexes()[op.Name]; found {
		return fmt.Errorf("duplicate index name: %s", op.Name)
	}
	indexes[op.Name] = op.Fields
	options := gomodels.Options{Table: op.table, Indexes: indexes}
	state.models[op.Model] = gomodels.New(
		model.Name(), model.Fields(), options,
	).Model
	return nil
}

func (op AddIndex) Run(tx *gomodels.Transaction, state *AppState) error {
	fields := state.models[op.Model].Fields()
	columns := make([]string, 0, len(op.Fields))
	for _, name := range op.Fields {
		column := fields[name].DBColumn(name)
		columns = append(columns, fmt.Sprintf("\"%s\"", column))
	}
	return tx.AddIndex(op.table, op.Name, columns...)
}

func (op AddIndex) Backwards(tx *gomodels.Transaction, pS *AppState) error {
	return tx.DropIndex(op.table, op.Name)
}

type RemoveIndex struct {
	Model string
	table string `json:"-"`
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
	op.table = model.Table()
	indexes := model.Indexes()
	if _, ok := model.Indexes()[op.Name]; !ok {
		return fmt.Errorf("index not found: %s", op.Name)
	}
	delete(indexes, op.Name)
	options := gomodels.Options{Table: op.table, Indexes: indexes}
	state.models[op.Model] = gomodels.New(
		model.Name(), model.Fields(), options,
	).Model
	return nil
}

func (op RemoveIndex) Run(tx *gomodels.Transaction, state *AppState) error {
	return tx.DropIndex(op.table, op.Name)
}

func (op RemoveIndex) Backwards(tx *gomodels.Transaction, pS *AppState) error {
	indexes := pS.models[op.Model].Indexes()
	fields := pS.models[op.Model].Fields()
	columns := make([]string, 0, len(indexes[op.Name]))
	for _, fieldName := range indexes[op.Name] {
		column := fields[fieldName].DBColumn(fieldName)
		columns = append(columns, fmt.Sprintf("\"%s\"", column))
	}
	return tx.AddIndex(op.table, op.Name, columns...)
}
