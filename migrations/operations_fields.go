package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"strings"
)

type AddFields struct {
	Model  string
	table  string `json:"-"`
	Fields gomodels.Fields
}

func (op AddFields) OpName() string {
	return "AddFields"
}

func (op AddFields) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return &op, err
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

func (op AddFields) Run(tx *sql.Tx, app string, driver string) error {
	baseQuery := fmt.Sprintf("ALTER TABLE \"%s\"", op.table)
	columns := make([]string, 0, len(op.Fields))
	for name, field := range op.Fields {
		addColumn := fmt.Sprintf(
			"ADD COLUMN \"%s\" %s",
			field.DBColumn(name), field.SqlDatatype(driver),
		)
		if driver == "sqlite3" {
			query := fmt.Sprintf("%s %s", baseQuery, addColumn)
			if _, err := tx.Exec(query); err != nil {
				return err
			}
		} else {
			columns = append(columns, addColumn)
		}
	}
	if driver == "postgres" {
		query := fmt.Sprintf("%s %s", baseQuery, strings.Join(columns, ", "))
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (op AddFields) Backwards(
	tx *sql.Tx, app string, driver string, pS *AppState,
) error {
	if driver == "postgres" {
		columns := make([]string, 0, len(op.Fields))
		for name := range op.Fields {
			columns = append(columns, fmt.Sprintf("DROP COLUMN %s", name))
		}
		query := fmt.Sprintf(
			"ALTER TABLE %s %s", op.table, strings.Join(columns, ", "),
		)
		if _, err := tx.Exec(query); err != nil {
			return err
		}
		return nil
	}
	query := fmt.Sprintf(
		"ALTER TABLE \"%[1]s\" RENAME TO \"%[1]s__old\"", op.table,
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	fields := pS.models[op.Model].Fields()
	columns := make([]string, 0, len(fields)-len(op.Fields))
	for name, field := range fields {
		columns = append(columns, fmt.Sprintf("\"%s\"", field.DBColumn(name)))
	}
	createModel := CreateModel{Name: op.Model, Fields: fields, Table: op.table}
	if err := createModel.Run(tx, app, driver); err != nil {
		return err
	}
	query = fmt.Sprintf(
		"INSERT INTO \"%[1]s\" (%[2]s) SELECT %[2]s FROM \"%[1]s__old\"",
		op.table, strings.Join(columns, ", "),
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	query = fmt.Sprintf("DROP TABLE \"%s__old\"", op.table)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	for idxName, fields := range pS.models[op.Model].Indexes() {
		addIndex := AddIndex{
			Model:  op.Model,
			Name:   idxName,
			Fields: fields,
			table:  op.table,
		}
		if err := addIndex.Run(tx, app, driver); err != nil {
			return err
		}
	}
	return nil
}

type RemoveFields struct {
	Model  string
	table  string `json:"-"`
	Fields []string
}

func (op RemoveFields) OpName() string {
	return "RemoveFields"
}

func (op RemoveFields) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return &op, err
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

func (op RemoveFields) Run(tx *sql.Tx, app string, driver string) error {
	if driver == "postgres" {
		columns := make([]string, 0, len(op.Fields))
		for _, name := range op.Fields {
			columns = append(columns, fmt.Sprintf("DROP COLUMN %s", name))
		}
		query := fmt.Sprintf(
			"ALTER TABLE %s %s", op.table, strings.Join(columns, ", "),
		)
		if _, err := tx.Exec(query); err != nil {
			return err
		}
		return nil
	}
	query := fmt.Sprintf(
		"ALTER TABLE \"%[1]s\" RENAME TO \"%[1]s__old\"", op.table,
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	fields := history[app].models[op.Model].Fields()
	keepColumns := make([]string, 0, len(fields)-len(op.Fields))
	for _, name := range op.Fields {
		delete(fields, name)
	}
	for name, field := range fields {
		keepColumns = append(
			keepColumns, fmt.Sprintf("\"%s\"", field.DBColumn(name)),
		)
	}
	createModel := CreateModel{Name: op.Model, Fields: fields, Table: op.table}
	if err := createModel.Run(tx, app, driver); err != nil {
		return err
	}
	query = fmt.Sprintf(
		"INSERT INTO \"%[1]s\" (%[2]s) SELECT %[2]s FROM \"%[1]s__old\"",
		op.table, strings.Join(keepColumns, ", "),
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	query = fmt.Sprintf("DROP TABLE \"%s__old\"", op.table)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	for idxName, fields := range history[app].models[op.Model].Indexes() {
		addIndex := AddIndex{
			Model:  op.Model,
			Name:   idxName,
			Fields: fields,
			table:  op.table,
		}
		if err := addIndex.Run(tx, app, driver); err != nil {
			return err
		}
	}
	return nil
}

func (op RemoveFields) Backwards(
	tx *sql.Tx, app string, driver string, pS *AppState,
) error {
	baseQuery := fmt.Sprintf("ALTER TABLE \"%s\"", op.table)
	fields := pS.models[op.Model].Fields()
	newFields := gomodels.Fields{}
	for _, name := range op.Fields {
		newFields[name] = fields[name]
	}
	columns := make([]string, 0, len(op.Fields))
	for name, field := range newFields {
		addColumn := fmt.Sprintf(
			"ADD COLUMN \"%s\" %s",
			field.DBColumn(name), field.SqlDatatype(driver),
		)
		if driver == "sqlite3" {
			query := fmt.Sprintf("%s %s", baseQuery, addColumn)
			if _, err := tx.Exec(query); err != nil {
				return err
			}
		} else {
			columns = append(columns, addColumn)
		}
	}
	if driver == "postgres" {
		query := fmt.Sprintf("%s %s", baseQuery, strings.Join(columns, ", "))
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}
	return nil
}
