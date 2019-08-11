package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

func testAddFieldOperationStateNoModel(t *testing.T, state *AppState) {
	op := AddFields{
		Model: "Customer",
		Fields: gomodels.Fields{
			"name": gomodels.CharField{},
		},
	}
	if err := op.SetState(state); err == nil {
		t.Errorf("expected model not found error")
	}
}

func testDuplicateFieldOperationState(t *testing.T, state *AppState) {
	op := AddFields{
		Model: "User",
		Fields: gomodels.Fields{
			"email": gomodels.CharField{},
		},
	}
	if err := op.SetState(state); err == nil {
		t.Errorf("expected duplicate field error")
	}
}

func testAddFieldOperationState(t *testing.T, state *AppState) {
	op := AddFields{
		Model: "User",
		Fields: gomodels.Fields{
			"firstName": gomodels.CharField{MaxLength: 50},
			"dob":       gomodels.DateField{},
		},
	}
	if err := op.SetState(state); err != nil {
		t.Errorf("%s", err)
	}
	fields := state.models["User"].Fields()
	if _, ok := fields["firstName"]; !ok {
		t.Errorf("state missing field firstName")
	}
	if _, ok := fields["dob"]; !ok {
		t.Errorf("state missing field dob")
	}
}

func testRemoveFieldOperationStateNoModel(t *testing.T, state *AppState) {
	op := RemoveFields{
		Model:  "Customer",
		Fields: []string{"name"},
	}
	if err := op.SetState(state); err == nil {
		t.Errorf("expected model not found error")
	}
}

func testRemoveFieldOperationStateNoField(t *testing.T, state *AppState) {
	op := RemoveFields{
		Model:  "User",
		Fields: []string{"lastName"},
	}
	if err := op.SetState(state); err == nil {
		t.Errorf("expected field not found error")
	}
}

func testRemoveFieldOperationState(t *testing.T, state *AppState) {
	op := RemoveFields{
		Model:  "User",
		Fields: []string{"email"},
	}
	if err := op.SetState(state); err != nil {
		t.Errorf("%s", err)
	}
	fields := state.models["User"].Fields()
	if _, ok := fields["email"]; ok {
		t.Errorf("email field was not removed from state")
	}
}

func testAddFieldOperationRunError(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	mockedEngine.Results.AddColumns = fmt.Errorf("db error")
	op := AddFields{
		Model: "User",
		Fields: gomodels.Fields{
			"firstName": gomodels.CharField{},
		},
	}
	fields := prevState.models["User"].Fields()
	fields["firstName"] = gomodels.CharField{}
	model := gomodels.New(
		"User", fields, gomodels.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Run(tx, state, prevState); err == nil {
		t.Errorf("expected db error")
	}
}

func testAddFieldOperationRun(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := AddFields{
		Model: "User",
		Fields: gomodels.Fields{
			"firstName": gomodels.CharField{},
		},
	}
	fields := prevState.models["User"].Fields()
	fields["firstName"] = gomodels.CharField{}
	model := gomodels.New(
		"User", fields, gomodels.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Run(tx, state, prevState); err != nil {
		t.Errorf("%s", err)
	}
	if mockedEngine.Calls("AddColumns") != 1 {
		t.Errorf("expected engine AddColumns to be called")
	}
}

func testAddFieldOperationBackwardError(
	t *testing.T,
	tx *gomodels.Transaction,
	state *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	mockedEngine.Results.DropColumns = fmt.Errorf("db error")
	op := AddFields{
		Model: "User",
		Fields: gomodels.Fields{
			"firstName": gomodels.CharField{},
		},
	}
	fields := state.models["User"].Fields()
	fields["firstName"] = gomodels.CharField{}
	model := gomodels.New(
		"User", fields, gomodels.Options{},
	).Model
	prevState := &AppState{
		app: state.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Backwards(tx, state, prevState); err == nil {
		t.Errorf("expected db error")
	}
}

func testAddFieldOperationBackwards(
	t *testing.T,
	tx *gomodels.Transaction,
	state *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := AddFields{
		Model: "User",
		Fields: gomodels.Fields{
			"firstName": gomodels.CharField{},
		},
	}
	fields := state.models["User"].Fields()
	fields["firstName"] = gomodels.CharField{}
	model := gomodels.New(
		"User", fields, gomodels.Options{},
	).Model
	prevState := &AppState{
		app: state.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Backwards(tx, state, prevState); err != nil {
		t.Errorf("%s", err)
	}
	if mockedEngine.Calls("DropColumns") != 1 {
		t.Errorf("expected engine DropColumns to be called")
	}
}

func testRemoveFieldOperationRunError(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	mockedEngine.Results.DropColumns = fmt.Errorf("db error")
	op := RemoveFields{
		Model:  "User",
		Fields: []string{"email"},
	}
	model := gomodels.New(
		"User", gomodels.Fields{}, gomodels.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Run(tx, state, prevState); err == nil {
		t.Errorf("expected db error")
	}
}

func testRemoveFieldOperationRun(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := RemoveFields{
		Model:  "User",
		Fields: []string{"email"},
	}
	model := gomodels.New(
		"User", gomodels.Fields{}, gomodels.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Run(tx, state, prevState); err != nil {
		t.Errorf("%s", err)
	}
	if mockedEngine.Calls("DropColumns") != 1 {
		t.Errorf("expected engine DropColumns to be called")
	}
}

func testRemoveFieldOperationBackwardError(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	mockedEngine.Results.AddColumns = fmt.Errorf("db error")
	op := RemoveFields{
		Model:  "User",
		Fields: []string{"email"},
	}
	model := gomodels.New(
		"User", gomodels.Fields{}, gomodels.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Backwards(tx, state, prevState); err == nil {
		t.Errorf("expected db error")
	}
}

func testRemoveFieldOperationBackwards(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := RemoveFields{
		Model:  "User",
		Fields: []string{"email"},
	}
	model := gomodels.New(
		"User", gomodels.Fields{}, gomodels.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	if err := op.Backwards(tx, state, prevState); err != nil {
		t.Errorf("%s", err)
	}
	if mockedEngine.Calls("AddColumns") != 1 {
		t.Errorf("expected engine DropColumns to be called")
	}
}

func TestFieldOperationsState(t *testing.T) {
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{
				MaxLength: 100, Index: true, DefaultEmpty: true,
			},
		},
		gomodels.Options{},
	)
	app := gomodels.NewApp("test", "", user.Model)
	if err := gomodels.Register(app); err != nil {
		panic(err)
	}
	defer gomodels.ClearRegistry()
	appState := &AppState{
		app: gomodels.Registry()["test"],
		models: map[string]*gomodels.Model{
			"User": user.Model,
		},
	}
	history["test"] = appState
	defer clearHistory()
	t.Run("AddFieldNoModel", func(t *testing.T) {
		testAddFieldOperationStateNoModel(t, appState)
	})
	t.Run("DuplicateField", func(t *testing.T) {
		testDuplicateFieldOperationState(t, appState)
	})
	t.Run("AddField", func(t *testing.T) {
		testAddFieldOperationState(t, appState)
	})
	t.Run("RemoveFieldNoModel", func(t *testing.T) {
		testRemoveFieldOperationStateNoModel(t, appState)
	})
	t.Run("RemoveNonExistentField", func(t *testing.T) {
		testRemoveFieldOperationStateNoField(t, appState)
	})
	t.Run("RemoveField", func(t *testing.T) {
		testRemoveFieldOperationState(t, appState)
	})
}

func TestFieldOperations(t *testing.T) {
	type funcType = func(t *testing.T, tx *gomodels.Transaction, s *AppState)
	matrix := map[string]funcType{
		"AddFieldRunError":         testAddFieldOperationRunError,
		"AddFieldRun":              testAddFieldOperationRun,
		"AddFieldBackwardError":    testAddFieldOperationBackwardError,
		"AddFieldBackwards":        testAddFieldOperationBackwards,
		"RemoveFieldRunError":      testRemoveFieldOperationRunError,
		"RemoveFieldRun":           testRemoveFieldOperationRun,
		"RemoveFieldBackwardError": testRemoveFieldOperationBackwardError,
		"RemoveFieldBackwards":     testRemoveFieldOperationBackwards,
	}
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{
				MaxLength: 100, Index: true, DefaultEmpty: true,
			},
		},
		gomodels.Options{},
	)
	app := gomodels.NewApp("test", "", user.Model)
	if err := gomodels.Register(app); err != nil {
		panic(err)
	}
	defer gomodels.ClearRegistry()
	appState := &AppState{
		app: gomodels.Registry()["test"],
		models: map[string]*gomodels.Model{
			"User": user.Model,
		},
	}
	history["test"] = appState
	defer clearHistory()
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		panic(err)
	}
	db := gomodels.Databases()["default"]
	tx, err := db.BeginTx()
	if err != nil {
		panic(err)
	}
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	defer gomodels.Stop()
	for name, f := range matrix {
		t.Run(name, func(t *testing.T) { f(t, tx, appState) })
		mockedEngine.Reset()
	}
}
