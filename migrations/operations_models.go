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
	Fields gomodels.Fields
}

func (op CreateModel) OpName() string {
	return "CreateModel"
}

func (op CreateModel) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op CreateModel) SetState(state *AppState) error {
	if _, found := state.Models[op.Name]; found {
		return fmt.Errorf("duplicate model: %s", op.Name)
	}
	state.Models[op.Name] = gomodels.New(
		op.Name, op.Fields, gomodels.Options{},
	).Model
	return nil
}

func (op CreateModel) Run(tx *sql.Tx, app string) error {
	query := fmt.Sprintf("CREATE TABLE '%s_%s' (", app, op.Name)
	fields := make([]string, 0, len(op.Fields))
	for name, field := range op.Fields {
		fields = append(
			fields,
			fmt.Sprintf("'%s' %s", field.DBColumn(name), field.CreateSQL()),
		)
	}
	query += strings.Join(fields, ", ") + ");"
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op CreateModel) Backwards(tx *sql.Tx, app string, pS *AppState) error {
	query := fmt.Sprintf("DROP TABLE '%s_%s';", app, op.Name)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

type DeleteModel struct {
	Name string
}

func (op DeleteModel) OpName() string {
	return "DeleteModel"
}

func (op DeleteModel) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op DeleteModel) SetState(state *AppState) error {
	if _, ok := state.Models[op.Name]; !ok {
		return fmt.Errorf("model not found: %s", op.Name)
	}
	delete(state.Models, op.Name)
	return nil
}

func (op DeleteModel) Run(tx *sql.Tx, app string) error {
	query := fmt.Sprintf("DROP TABLE '%s_%s';", app, op.Name)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op DeleteModel) Backwards(tx *sql.Tx, app string, pS *AppState) error {
	model := pS.Models[op.Name]
	query := fmt.Sprintf("CREATE TABLE '%s_%s' (", app, model.Name())
	fields := make([]string, 0, len(model.Fields()))
	for name, field := range model.Fields() {
		fields = append(
			fields,
			fmt.Sprintf("'%s' %s", field.DBColumn(name), field.CreateSQL()),
		)
	}
	query += strings.Join(fields, ", ") + ");"
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

type AddIndex struct {
	Model  string
	Name   string
	Fields []string
}

func (op AddIndex) OpName() string {
	return "AddIndex"
}

func (op AddIndex) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op AddIndex) SetState(state *AppState) error {
	model, ok := state.Models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	indexes := model.Indexes()
	if _, found := model.Indexes()[op.Name]; found {
		return fmt.Errorf("duplicate index name: %s", op.Name)
	}
	indexes[op.Name] = op.Fields
	state.Models[op.Model] = gomodels.New(
		model.Name(), model.Fields(), gomodels.Options{Indexes: indexes},
	).Model
	return nil
}

func (op AddIndex) Run(tx *sql.Tx, app string) error {
	query := fmt.Sprintf(
		"CREATE INDEX '%s' ON '%s_%s'", op.Name, app, op.Model,
	)
	fields := history[app].Models[op.Model].Fields()
	columns := make([]string, 0, len(op.Fields))
	for _, name := range op.Fields {
		column := fields[name].DBColumn(name)
		columns = append(columns, fmt.Sprintf("'%s'", column))
	}
	query += fmt.Sprintf(" (%s);", strings.Join(columns, ", "))
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op AddIndex) Backwards(tx *sql.Tx, app string, pS *AppState) error {
	query := fmt.Sprintf("DROP INDEX '%s';", op.Name)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

type RemoveIndex struct {
	Model string
	Name  string
}

func (op RemoveIndex) OpName() string {
	return "RemoveIndex"
}

func (op RemoveIndex) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op RemoveIndex) SetState(state *AppState) error {
	model, ok := state.Models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	indexes := model.Indexes()
	if _, ok := model.Indexes()[op.Name]; !ok {
		return fmt.Errorf("index not found: %s", op.Name)
	}
	delete(indexes, op.Name)
	state.Models[op.Model] = gomodels.New(
		model.Name(), model.Fields(), gomodels.Options{Indexes: indexes},
	).Model
	return nil
}

func (op RemoveIndex) Run(tx *sql.Tx, app string) error {
	query := fmt.Sprintf(
		"DROP INDEX '%s' ON '%s_%s';", op.Name, app, op.Model,
	)
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}

func (op RemoveIndex) Backwards(tx *sql.Tx, app string, pS *AppState) error {
	model := pS.Models[op.Model]
	indexes := model.Indexes()
	fields := model.Fields()
	query := fmt.Sprintf(
		"CREATE INDEX '%s' ON '%s_%s'", op.Name, app, op.Model,
	)
	columns := make([]string, 0, len(indexes[op.Name]))
	for _, fieldName := range indexes[op.Name] {
		column := fields[fieldName].DBColumn(fieldName)
		columns = append(columns, fmt.Sprintf("'%s'", column))
	}
	query += fmt.Sprintf(" (%s);", strings.Join(columns, ", "))
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	return nil
}
