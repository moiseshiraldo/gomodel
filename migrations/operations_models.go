package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
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

func (op CreateModel) Run(tx *sql.Tx) error {
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

func (op DeleteModel) Run(tx *sql.Tx) error {
	return nil
}
