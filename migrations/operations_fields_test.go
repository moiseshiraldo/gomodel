package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

func TestFieldOperationsState(t *testing.T) {
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{MaxLength: 100, Index: true},
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
		op := AddFields{
			Model: "Customer",
			Fields: gomodels.Fields{
				"name": gomodels.CharField{},
			},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected model not found error")
		}
	})
	t.Run("DuplicateField", func(t *testing.T) {
		op := AddFields{
			Model: "User",
			Fields: gomodels.Fields{
				"email": gomodels.CharField{},
			},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected duplicate field error")
		}
	})
	t.Run("AddField", func(t *testing.T) {
		op := AddFields{
			Model: "User",
			Fields: gomodels.Fields{
				"firstName": gomodels.CharField{MaxLength: 50},
				"dob":       gomodels.DateField{},
			},
		}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		fields := appState.models["User"].Fields()
		if _, ok := fields["firstName"]; !ok {
			t.Errorf("state missing field firstName")
		}
		if _, ok := fields["dob"]; !ok {
			t.Errorf("state missing field dob")
		}
	})
	t.Run("RemoveFieldNoModel", func(t *testing.T) {
		op := RemoveFields{
			Model:  "Customer",
			Fields: []string{"name"},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected model not found error")
		}
	})
	t.Run("RemoveNonExistentField", func(t *testing.T) {
		op := RemoveFields{
			Model:  "User",
			Fields: []string{"lastName"},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected field not found error")
		}
	})
	t.Run("RemoveField", func(t *testing.T) {
		op := RemoveFields{
			Model:  "User",
			Fields: []string{"email"},
		}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		fields := appState.models["User"].Fields()
		if _, ok := fields["email"]; ok {
			t.Errorf("email field was not removed from state")
		}
	})
}

func TestFieldOperations(t *testing.T) {
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{MaxLength: 100, Index: true},
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
	defer gomodels.Stop()
	db := gomodels.Databases()["default"]
	tx, err := db.BeginTx()
	if err != nil {
		panic(err)
	}
	t.Run("AddField", func(t *testing.T) {
		testAddFieldOperation(t, tx, appState)
	})
	t.Run("RemoveField", func(t *testing.T) {
		testRemoveFieldOperation(t, tx, appState)
	})
}

func testAddFieldOperation(
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
	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddColumns = fmt.Errorf("db error")
		if err := op.Run(tx, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(tx, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddColumns") != 1 {
			t.Errorf("expected engine AddColumns to be called")
		}
	})
	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropColumns = fmt.Errorf("db error")
		if err := op.Backwards(tx, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(tx, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropColumns") != 1 {
			t.Errorf("expected engine DropColumns to be called")
		}
	})
}

func testRemoveFieldOperation(
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
	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropColumns = fmt.Errorf("db error")
		if err := op.Run(tx, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(tx, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropColumns") != 1 {
			t.Errorf("expected engine DropColumns to be called")
		}
	})
	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddColumns = fmt.Errorf("db error")
		if err := op.Backwards(tx, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(tx, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddColumns") != 1 {
			t.Errorf("expected engine DropColumns to be called")
		}
	})
}
