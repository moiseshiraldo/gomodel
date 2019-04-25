package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"strings"
)

type CreateModel struct {
	Model  string
	Fields gomodels.Fields
}

func (op CreateModel) Name() string {
	return "CreateModel"
}

func (op CreateModel) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op CreateModel) SetState(state *AppState) error {
	if _, found := state.Models[op.Model]; found {
		return fmt.Errorf("duplicate model: %s", op.Model)
	}
	state.Models[op.Model] = gomodels.New(op.Model, op.Fields)
	return nil
}

func (op CreateModel) Run(tx *sql.Tx, app string) error {
	query := fmt.Sprintf("CREATE TABLE '%s_%s' (", app, op.Model)
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
	Model string
}

func (op DeleteModel) Name() string {
	return "DeleteModel"
}

func (op DeleteModel) FromJSON(raw []byte) (Operation, error) {
	err := json.Unmarshal(raw, &op)
	return op, err
}

func (op DeleteModel) SetState(state *AppState) error {
	if _, ok := state.Models[op.Model]; !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	delete(state.Models, op.Model)
	return nil
}

func (op DeleteModel) Run(tx *sql.Tx, app string) error {
	return nil
}
