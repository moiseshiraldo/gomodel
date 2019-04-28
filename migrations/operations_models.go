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
	state.Models[op.Name] = gomodels.New(op.Name, op.Fields)
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

type AddIndex struct {
	Model   string
	Name    string
	Columns []string
}

func (op AddIndex) OpName() string {
	return "AddIndex"
}

func (op AddIndex) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op AddIndex) SetState(state *AppState) error {
	return nil
}

func (op AddIndex) Run(tx *sql.Tx, app string) error {
	query := fmt.Sprintf(
		"CREATE INDEX '%s' ON '%s_%s'", op.Name, app, op.Model,
	)
	columns := make([]string, 0, len(op.Columns))
	for _, column := range op.Columns {
		columns = append(columns, fmt.Sprintf("'%s'", column))
	}
	query += fmt.Sprintf(" (%s);", strings.Join(columns, ", "))
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
