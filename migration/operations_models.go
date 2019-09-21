package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
)

// CreateModel implements the Operation interface to create a new model.
type CreateModel struct {
	Name string
	// Table is used to set a custom name for the database table. If blank,
	// the table will be created as {app_name}_{model_name} all lowercase.
	Table  string `json:",omitempty"`
	Fields gomodel.Fields
}

// OpName returns the operation name.
func (op CreateModel) OpName() string {
	return "CreateModel"
}

// SetState adds the new model to the given application state.
func (op CreateModel) SetState(state *AppState) error {
	if _, found := state.Models[op.Name]; found {
		return fmt.Errorf("duplicate model: %s", op.Name)
	}
	table := op.Table
	if table == "" {
		table = fmt.Sprintf("%s_%s", state.app.Name(), op.Name)
	}
	fields := gomodel.Fields{}
	for name, field := range op.Fields {
		fields[name] = field
	}
	state.Models[op.Name] = gomodel.New(
		op.Name, fields, gomodel.Options{Table: table},
	).Model
	return nil
}

// Run creates the table on the database.
func (op CreateModel) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.CreateTable(state.Models[op.Name], true)
}

// Backwards drops the table from the database.
func (op CreateModel) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropTable(state.Models[op.Name])
}

// DeleteModel implements the operation to delete a model.
type DeleteModel struct {
	Name string
}

// OpName returns the operation name.
func (op DeleteModel) OpName() string {
	return "DeleteModel"
}

// SetState removes the model from the given application state.
func (op DeleteModel) SetState(state *AppState) error {
	if _, ok := state.Models[op.Name]; !ok {
		return fmt.Errorf("model not found: %s", op.Name)
	}
	delete(state.Models, op.Name)
	return nil
}

// Run drops the table from the database.
func (op DeleteModel) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropTable(prevState.Models[op.Name])
}

// Backwards creates the table on the database.
func (op DeleteModel) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.CreateTable(prevState.Models[op.Name], true)
}

// AddIndex implements the Operation interface to add an index.
type AddIndex struct {
	Model  string
	Name   string
	Fields []string
}

// OpName returns the operation name.
func (op AddIndex) OpName() string {
	return "AddIndex"
}

// SetState adds the index to the model in the given application state.
func (op AddIndex) SetState(state *AppState) error {
	model, ok := state.Models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	return model.AddIndex(op.Name, op.Fields...)
}

// Run creates the index on the database.
func (op AddIndex) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.AddIndex(state.Models[op.Model], op.Name, op.Fields...)
}

// Backwards drops the index from the database.
func (op AddIndex) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropIndex(state.Models[op.Model], op.Name)
}

// RemoveIndex implements the Operation interface to remove an index.
type RemoveIndex struct {
	Model string
	Name  string
}

// OpName returns the operation name.
func (op RemoveIndex) OpName() string {
	return "RemoveIndex"
}

// SetState removes the index from the model in the given application state.
func (op RemoveIndex) SetState(state *AppState) error {
	model, ok := state.Models[op.Model]
	if !ok {
		return fmt.Errorf("model not found: %s", op.Model)
	}
	return model.RemoveIndex(op.Name)
}

// Run drops the index from the database.
func (op RemoveIndex) Run(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	return engine.DropIndex(state.Models[op.Model], op.Name)
}

// Backwards creates the index on the database.
func (op RemoveIndex) Backwards(
	engine gomodel.Engine,
	state *AppState,
	prevState *AppState,
) error {
	model := prevState.Models[op.Model]
	return engine.AddIndex(model, op.Name, model.Indexes()[op.Name]...)
}
