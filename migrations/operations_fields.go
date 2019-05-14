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
	Fields gomodels.Fields
}

func (op AddFields) OpName() string {
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
	state.Models[op.Model] = gomodels.New(
		op.Model, fields, gomodels.Options{},
	).Model
	return nil
}

func (op AddFields) Run(tx *sql.Tx, app string) error {
	baseQuery := fmt.Sprintf("ALTER TABLE '%s_%s' ADD COLUMN", app, op.Model)
	for name, field := range op.Fields {
		query := fmt.Sprintf(
			"%s '%s' %s;", baseQuery, field.DBColumn(name), field.CreateSQL(),
		)
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (op AddFields) Backwards(tx *sql.Tx, app string) error {
	query := fmt.Sprintf(
		"ALTER TABLE %[1]s_%[2]s RENAME TO %[1]s_%[2]s__old;", app, op.Model,
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	fields := history[app].Models[op.Model].Fields()
	keepList := make([]string, 0, len(fields)-len(op.Fields))
	for name := range op.Fields {
		delete(fields, name)
	}
	for name := range fields {
		keepList = append(keepList, name)
	}
	createModel := CreateModel{Name: op.Model, Fields: fields}
	if err := createModel.Run(tx, app); err != nil {
		return err
	}
	query = fmt.Sprintf(
		"INSERT INTO %[1]s_%[2]s (%[3]s) SELECT %[3]s FROM %[1]s_%[2]s__old;",
		app, op.Model, strings.Join(keepList, ", "),
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

type RemoveFields struct {
	Model  string
	Fields []string
}

func (op RemoveFields) OpName() string {
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
	state.Models[op.Model] = gomodels.New(
		op.Model, fields, gomodels.Options{},
	).Model
	return nil
}

func (op RemoveFields) Run(tx *sql.Tx, app string) error {
	query := fmt.Sprintf(
		"ALTER TABLE %[1]s_%[2]s RENAME TO %[1]s_%[2]s__old;", app, op.Model,
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	fields := history[app].Models[op.Model].Fields()
	keepList := make([]string, 0, len(fields)-len(op.Fields))
	for _, name := range op.Fields {
		delete(fields, name)
	}
	for name := range fields {
		keepList = append(keepList, name)
	}
	createModel := CreateModel{Name: op.Model, Fields: fields}
	if err := createModel.Run(tx, app); err != nil {
		return err
	}
	query = fmt.Sprintf(
		"INSERT INTO %[1]s_%[2]s (%[3]s) SELECT %[3]s FROM %[1]s_%[2]s__old;",
		app, op.Model, strings.Join(keepList, ", "),
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op RemoveFields) Backwards(tx *sql.Tx, app string) error {
	return nil
}
