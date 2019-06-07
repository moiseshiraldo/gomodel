package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"strings"
)

type CreateModel struct {
	Name   string
	Table  string `json:",omitempty"`
	Fields gomodels.Fields
}

func (op CreateModel) OpName() string {
	return "CreateModel"
}

func (op CreateModel) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return &op, err
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

func (op CreateModel) Run(tx *sql.Tx, app string, driver string) error {
	if op.Table == "" {
		op.Table = fmt.Sprintf("%s_%s", app, op.Name)
	}
	query := fmt.Sprintf("CREATE TABLE \"%s\" (", op.Table)
	fields := make([]string, 0, len(op.Fields))
	for name, field := range op.Fields {
		sqlColumn := fmt.Sprintf(
			"\"%s\" %s", field.DBColumn(name), field.SqlDatatype(driver),
		)
		fields = append(fields, sqlColumn)
	}
	query += strings.Join(fields, ", ") + ")"
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op CreateModel) Backwards(
	tx *sql.Tx, app string, driver string, pS *AppState,
) error {
	if op.Table == "" {
		op.Table = fmt.Sprintf("%s_%s", app, op.Name)
	}
	query := fmt.Sprintf("DROP TABLE \"%s\"", op.Table)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

type DeleteModel struct {
	Name  string
	table string `json:"-"`
}

func (op DeleteModel) OpName() string {
	return "DeleteModel"
}

func (op DeleteModel) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return &op, err
}

func (op *DeleteModel) SetState(state *AppState) error {
	if _, ok := state.models[op.Name]; !ok {
		return fmt.Errorf("model not found: %s", op.Name)
	}
	op.table = state.models[op.Name].Table()
	delete(state.models, op.Name)
	return nil
}

func (op DeleteModel) Run(tx *sql.Tx, app string, driver string) error {
	query := fmt.Sprintf("DROP TABLE \"%s\"", op.table)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op DeleteModel) Backwards(
	tx *sql.Tx, app string, driver string, pS *AppState,
) error {
	model := pS.models[op.Name]
	query := fmt.Sprintf("CREATE TABLE \"%s\" (", model.Table())
	fields := make([]string, 0, len(model.Fields()))
	for name, field := range model.Fields() {
		sqlColumn := fmt.Sprintf(
			"\"%s\" %s", field.DBColumn(name), field.SqlDatatype(driver),
		)
		fields = append(fields, sqlColumn)
	}
	query += strings.Join(fields, ", ") + ");"
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
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

func (op AddIndex) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return &op, err
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

func (op AddIndex) Run(tx *sql.Tx, app string, driver string) error {
	query := fmt.Sprintf(
		"CREATE INDEX \"%s\" ON \"%s\"", op.Name, op.table,
	)
	fields := history[app].models[op.Model].Fields()
	columns := make([]string, 0, len(op.Fields))
	for _, name := range op.Fields {
		column := fields[name].DBColumn(name)
		columns = append(columns, fmt.Sprintf("\"%s\"", column))
	}
	query += fmt.Sprintf(" (%s)", strings.Join(columns, ", "))
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op AddIndex) Backwards(
	tx *sql.Tx, app string, driver string, pS *AppState,
) error {
	query := fmt.Sprintf("DROP INDEX \"%s\"", op.Name)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

type RemoveIndex struct {
	Model string
	table string `json:"-"`
	Name  string
}

func (op RemoveIndex) OpName() string {
	return "RemoveIndex"
}

func (op RemoveIndex) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return &op, err
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

func (op RemoveIndex) Run(tx *sql.Tx, app string, driver string) error {
	query := fmt.Sprintf(
		"DROP INDEX \"%s\" ON \"%s\"", op.Name, op.table,
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op RemoveIndex) Backwards(
	tx *sql.Tx, app string, driver string, pS *AppState,
) error {
	model := pS.models[op.Model]
	indexes := model.Indexes()
	fields := model.Fields()
	query := fmt.Sprintf(
		"CREATE INDEX \"%s\" ON \"%s\"", op.Name, model.Table(),
	)
	columns := make([]string, 0, len(indexes[op.Name]))
	for _, fieldName := range indexes[op.Name] {
		column := fields[fieldName].DBColumn(fieldName)
		columns = append(columns, fmt.Sprintf("\"%s\"", column))
	}
	query += fmt.Sprintf(" (%s)", strings.Join(columns, ", "))
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}
