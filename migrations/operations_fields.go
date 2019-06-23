package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
)

type AddFields struct {
	Model  string
	table  string `json:"-"`
	Fields gomodels.Fields
}

func (op AddFields) OpName() string {
	return "AddFields"
}

func (op *AddFields) SetState(state *AppState) error {
	if _, ok := state.models[op.Model]; !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	op.table = state.models[op.Model].Table()
	fields := state.models[op.Model].Fields()
	for name, field := range op.Fields {
		if _, found := fields[name]; found {
			return fmt.Errorf("%s: duplicate field: %s", op.Model, name)
		}
		fields[name] = field
	}
	options := gomodels.Options{
		Table: op.table, Indexes: state.models[op.Model].Indexes(),
	}
	delete(state.models, op.Model)
	state.models[op.Model] = gomodels.New(
		op.Model, fields, options,
	).Model
	return nil
}

func (op AddFields) Run(tx *gomodels.Transaction, state *AppState) error {
	return tx.AddColumns(op.table, op.Fields)
}

func (op AddFields) Backwards(tx *gomodels.Transaction, pS *AppState) error {
	if _, ok := tx.Engine.(gomodels.SqliteEngine); ok {
		fields := pS.models[op.Model].Fields()
		keepCols := make([]string, 0, len(fields)-len(op.Fields))
		for name, field := range fields {
			keepCols = append(keepCols, field.DBColumn(name))
		}
		name := op.table + "__new"
		if err := tx.CopyTable(op.table, name, keepCols...); err != nil {
			return err
		}
		if err := tx.DropTable(op.table); err != nil {
			return err
		}
		if err := tx.RenameTable(name, op.table); err != nil {
			return err
		}
		for idxName, fields := range pS.models[op.Model].Indexes() {
			addIndex := AddIndex{
				Model:  op.Model,
				Name:   idxName,
				Fields: fields,
				table:  op.table,
			}
			if err := addIndex.Run(tx, history[pS.app.Name()]); err != nil {
				return err
			}
		}
		return nil
	}
	columns := make([]string, 0, len(op.Fields))
	for name, field := range op.Fields {
		columns = append(columns, field.DBColumn(name))
	}
	return tx.DropColumns(op.table, columns...)
}

type RemoveFields struct {
	Model  string
	table  string `json:"-"`
	Fields []string
}

func (op RemoveFields) OpName() string {
	return "RemoveFields"
}

func (op *RemoveFields) SetState(state *AppState) error {
	if _, ok := state.models[op.Model]; !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	op.table = state.models[op.Model].Table()
	fields := state.models[op.Model].Fields()
	for _, name := range op.Fields {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("%s: field not found: %s", op.Model, name)
		}
		delete(fields, name)
	}
	options := gomodels.Options{
		Table: op.table, Indexes: state.models[op.Model].Indexes(),
	}
	delete(state.models, op.Model)
	state.models[op.Model] = gomodels.New(
		op.Model, fields, options,
	).Model
	return nil
}

func (op RemoveFields) Run(tx *gomodels.Transaction, state *AppState) error {
	if _, ok := tx.Engine.(gomodels.SqliteEngine); ok {
		fields := state.models[op.Model].Fields()
		keepCols := make([]string, 0, len(fields)-len(op.Fields))
		for _, name := range op.Fields {
			delete(fields, name)
		}
		for name, field := range fields {
			keepCols = append(keepCols, field.DBColumn(name))
		}
		name := op.table + "__new"
		if err := tx.CopyTable(op.table, name, keepCols...); err != nil {
			return err
		}
		if err := tx.DropTable(op.table); err != nil {
			return err
		}
		if err := tx.RenameTable(name, op.table); err != nil {
			return err
		}
		for idxName, fields := range state.models[op.Model].Indexes() {
			addIndex := AddIndex{
				Model:  op.Model,
				Name:   idxName,
				Fields: fields,
				table:  op.table,
			}
			if err := addIndex.Run(tx, state); err != nil {
				return err
			}
		}
		return nil
	}
	return tx.DropColumns(op.table, op.Fields...)
}

func (op RemoveFields) Backwards(tx *gomodels.Transaction, pS *AppState) error {
	fields := pS.models[op.Model].Fields()
	newFields := gomodels.Fields{}
	for _, name := range op.Fields {
		newFields[name] = fields[name]
	}
	return tx.AddColumns(op.table, newFields)
}
