package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"testing"
)

// TestFieldOperationsState tests field operations SetState method
func TestFieldOperationsState(t *testing.T) {
	// Models setup
	user := gomodel.New(
		"User",
		gomodel.Fields{
			"email":         gomodel.CharField{MaxLength: 100, Index: true},
			"loginAttempts": gomodel.IntegerField{DefaultZero: true},
		},
		gomodel.Options{},
	)
	// App setup
	app := gomodel.NewApp("test", "", user.Model)
	gomodel.Register(app)
	defer gomodel.ClearRegistry()
	// App state setup
	appState := &AppState{
		app: gomodel.Registry()["test"],
		Models: map[string]*gomodel.Model{
			"User": user.Model,
		},
	}
	history["test"] = appState
	defer clearHistory()

	t.Run("AddFieldNoModel", func(t *testing.T) {
		op := AddFields{
			Model: "Customer",
			Fields: gomodel.Fields{
				"name": gomodel.CharField{},
			},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected model not found error")
		}
	})

	t.Run("DuplicateField", func(t *testing.T) {
		op := AddFields{
			Model: "User",
			Fields: gomodel.Fields{
				"email": gomodel.CharField{},
			},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected duplicate field error")
		}
	})

	t.Run("AddField", func(t *testing.T) {
		op := AddFields{
			Model: "User",
			Fields: gomodel.Fields{
				"firstName": gomodel.CharField{MaxLength: 50},
				"dob":       gomodel.DateField{},
			},
		}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		fields := appState.Models["User"].Fields()
		if _, ok := fields["firstName"]; !ok {
			t.Errorf("model state is missing field firstName")
		}
		if _, ok := fields["dob"]; !ok {
			t.Errorf("model state is missing field dob")
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

	t.Run("RemoveUnknowntField", func(t *testing.T) {
		op := RemoveFields{
			Model:  "User",
			Fields: []string{"lastName"},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected field not found error")
		}
	})

	t.Run("RemoveIndexedField", func(t *testing.T) {
		op := RemoveFields{
			Model:  "User",
			Fields: []string{"email"},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected cannot remove indexed field error")
		}
	})

	t.Run("RemoveField", func(t *testing.T) {
		op := RemoveFields{
			Model:  "User",
			Fields: []string{"loginAttempts"},
		}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		fields := appState.Models["User"].Fields()
		if _, ok := fields["loginAttemptsl"]; ok {
			t.Errorf("email field was not removed from model state")
		}
	})
}

// TestFieldOperations tests field operations Run/Backwards methods
func TestFieldOperations(t *testing.T) {
	// Models setup
	user := gomodel.New(
		"User",
		gomodel.Fields{
			"email": gomodel.CharField{MaxLength: 100, Index: true},
		},
		gomodel.Options{},
	)
	// App setup
	app := gomodel.NewApp("test", "", user.Model)
	gomodel.Register(app)
	defer gomodel.ClearRegistry()
	// App state setup
	appState := &AppState{
		app: gomodel.Registry()["test"],
		Models: map[string]*gomodel.Model{
			"User": user.Model,
		},
	}
	history["test"] = appState
	defer clearHistory()
	// DB setup
	err := gomodel.Start(map[string]gomodel.Database{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gomodel.Stop()
	db := gomodel.Databases()["default"]
	engine := db.Engine.(gomodel.MockedEngine)
	t.Run("AddField", func(t *testing.T) {
		testAddFieldOperation(t, engine, appState)
	})
	t.Run("RemoveField", func(t *testing.T) {
		testRemoveFieldOperation(t, engine, appState)
	})
}

func testAddFieldOperation(
	t *testing.T,
	mockedEngine gomodel.MockedEngine,
	prevState *AppState,
) {
	op := AddFields{
		Model: "User",
		Fields: gomodel.Fields{
			"firstName": gomodel.CharField{},
		},
	}
	fields := prevState.Models["User"].Fields()
	fields["firstName"] = gomodel.CharField{}
	model := gomodel.New(
		"User", fields, gomodel.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		Models: map[string]*gomodel.Model{
			"User": model,
		},
	}

	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddColumns = fmt.Errorf("db error")
		if err := op.Run(mockedEngine, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(mockedEngine, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddColumns") != 1 {
			t.Errorf("expected engine AddColumns to be called")
		}
	})

	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropColumns = fmt.Errorf("db error")
		if err := op.Backwards(mockedEngine, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(mockedEngine, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropColumns") != 1 {
			t.Errorf("expected engine DropColumns to be called")
		}
	})
}

func testRemoveFieldOperation(
	t *testing.T,
	mockedEngine gomodel.MockedEngine,
	prevState *AppState,
) {
	op := RemoveFields{
		Model:  "User",
		Fields: []string{"email"},
	}
	model := gomodel.New(
		"User", gomodel.Fields{}, gomodel.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		Models: map[string]*gomodel.Model{
			"User": model,
		},
	}

	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropColumns = fmt.Errorf("db error")
		if err := op.Run(mockedEngine, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(mockedEngine, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropColumns") != 1 {
			t.Errorf("expected engine DropColumns to be called")
		}
	})

	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddColumns = fmt.Errorf("db error")
		if err := op.Backwards(mockedEngine, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(mockedEngine, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddColumns") != 1 {
			t.Errorf("expected engine DropColumns to be called")
		}
	})
}
