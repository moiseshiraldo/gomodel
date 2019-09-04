package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"testing"
)

// TestModelOperationsState tests model operations SetState method
func TestModelOperationsState(t *testing.T) {
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

	t.Run("AddDuplicateModel", func(t *testing.T) {
		op := CreateModel{Name: "User"}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected duplicate model error")
		}
	})

	t.Run("AddModel", func(t *testing.T) {
		op := CreateModel{
			Name: "Customer",
			Fields: gomodel.Fields{
				"name": gomodel.CharField{MaxLength: 100},
			},
		}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		model, ok := appState.Models["Customer"]
		if !ok {
			t.Errorf("model was not added to state")
		} else {
			if _, ok := model.Fields()["name"]; !ok {
				t.Errorf("model state missing name field")
			}
		}
	})

	t.Run("AddDuplicateIndex", func(t *testing.T) {
		op := AddIndex{Model: "User", Name: "test_user_email_auto_idx"}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected duplicate index error")
		}
	})

	t.Run("AddIndexNoModel", func(t *testing.T) {
		op := AddIndex{Model: "Transaction", Name: "test_index"}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected model not found error")
		}
	})

	t.Run("AddIndexNoField", func(t *testing.T) {
		op := AddIndex{
			Model:  "User",
			Name:   "test_index",
			Fields: []string{"username"},
		}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected unknown field error")
		}
	})

	t.Run("AddIndex", func(t *testing.T) {
		op := AddIndex{
			Model:  "User",
			Name:   "test_index",
			Fields: []string{"email"},
		}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		model := appState.Models["User"]
		if _, ok := model.Indexes()["test_index"]; !ok {
			t.Errorf("index was not added to model state")
		}
	})

	t.Run("RemoveIndexNoModel", func(t *testing.T) {
		op := RemoveIndex{Model: "Transaction", Name: "test_index"}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected model not found error")
		}
	})

	t.Run("RemoveMissingIndex", func(t *testing.T) {
		op := RemoveIndex{Model: "User", Name: "missing_index"}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected index not found error")
		}
	})

	t.Run("RemoveIndex", func(t *testing.T) {
		op := RemoveIndex{Model: "User", Name: "test_user_email_auto_idx"}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		model := appState.Models["User"]
		if _, found := model.Indexes()["test_user_email_auto_idx"]; found {
			t.Errorf("index was not removed from model state")
		}
	})

	t.Run("RemoveMissingModel", func(t *testing.T) {
		op := DeleteModel{Name: "Transaction"}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected model not found error")
		}
	})

	t.Run("RemoveModel", func(t *testing.T) {
		op := DeleteModel{Name: "User"}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		if _, found := appState.Models["User"]; found {
			t.Errorf("model was not removed from state")
		}
	})
}

// TestModelOperations tests model operations Run/Backwards methods
func TestModelOperations(t *testing.T) {
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
	t.Run("AddModel", func(t *testing.T) {
		testAddModelOperation(t, engine, appState)
	})
	t.Run("DeleteModel", func(t *testing.T) {
		testDeleteModelOperation(t, engine, appState)
	})
	t.Run("AddIndex", func(t *testing.T) {
		testAddIndexOperation(t, engine, appState)
	})
	t.Run("RemoveIndex", func(t *testing.T) {
		testRemoveIndexOperation(t, engine, appState)
	})
}

func testAddModelOperation(
	t *testing.T,
	mockedEngine gomodel.MockedEngine,
	prevState *AppState,
) {
	op := CreateModel{Name: "Customer"}
	model := gomodel.New(
		"Customer", gomodel.Fields{}, gomodel.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		Models: map[string]*gomodel.Model{
			"User":     prevState.Models["User"],
			"Customer": model,
		},
	}

	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.CreateTable = fmt.Errorf("db error")
		if err := op.Run(mockedEngine, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(mockedEngine, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("CreateTable") != 1 {
			t.Errorf("expected engine CreateTable to be called")
		}
	})

	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropTable = fmt.Errorf("db error")
		if err := op.Backwards(mockedEngine, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(mockedEngine, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropTable") != 1 {
			t.Errorf("expected engine DropTable to be called")
		}
	})
}

func testDeleteModelOperation(
	t *testing.T,
	mockedEngine gomodel.MockedEngine,
	prevState *AppState,
) {
	op := DeleteModel{Name: "User"}
	state := &AppState{
		app:    prevState.app,
		Models: map[string]*gomodel.Model{},
	}

	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropTable = fmt.Errorf("db error")
		if err := op.Run(mockedEngine, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(mockedEngine, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropTable") != 1 {
			t.Errorf("expected engine DropTable to be called")
		}
	})

	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.CreateTable = fmt.Errorf("db error")
		if err := op.Backwards(mockedEngine, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(mockedEngine, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("CreateTable") != 1 {
			t.Errorf("expected engine CreateTable to be called")
		}
	})
}

func testAddIndexOperation(
	t *testing.T,
	mockedEngine gomodel.MockedEngine,
	prevState *AppState,
) {
	op := AddIndex{
		Model:  "User",
		Name:   "test_index",
		Fields: []string{"email"},
	}
	indexes := prevState.Models["User"].Indexes()
	indexes["test_index"] = []string{"email"}
	model := gomodel.New(
		"User",
		prevState.Models["User"].Fields(),
		gomodel.Options{Indexes: indexes},
	).Model
	state := &AppState{
		app: prevState.app,
		Models: map[string]*gomodel.Model{
			"User": model,
		},
	}

	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddIndex = fmt.Errorf("db error")
		if err := op.Run(mockedEngine, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(mockedEngine, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddIndex") != 1 {
			t.Errorf("expected engine AddIndex to be called")
		}
	})

	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropIndex = fmt.Errorf("db error")
		if err := op.Backwards(mockedEngine, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(mockedEngine, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropIndex") != 1 {
			t.Errorf("expected engine DropIndex to be called")
		}
	})
}

func testRemoveIndexOperation(
	t *testing.T,
	mockedEngine gomodel.MockedEngine,
	prevState *AppState,
) {
	op := RemoveIndex{
		Model: "User",
		Name:  "test_index",
	}
	indexes := prevState.Models["User"].Indexes()
	delete(indexes, "test_user_email_auto_idx")
	model := gomodel.New(
		"User",
		prevState.Models["User"].Fields(),
		gomodel.Options{Indexes: indexes},
	).Model
	state := &AppState{
		app: prevState.app,
		Models: map[string]*gomodel.Model{
			"User": model,
		},
	}

	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropIndex = fmt.Errorf("db error")
		if err := op.Run(mockedEngine, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(mockedEngine, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropIndex") != 1 {
			t.Errorf("expected engine DropIndex to be called")
		}
	})

	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddIndex = fmt.Errorf("db error")
		if err := op.Backwards(mockedEngine, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(mockedEngine, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddIndex") != 1 {
			t.Errorf("expected engine AddIndex to be called")
		}
	})
}
