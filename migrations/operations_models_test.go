package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

func TestModelOperationsState(t *testing.T) {
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{MaxLength: 100, Index: true},
		},
		gomodels.Options{},
	)
	app := gomodels.NewApp("test", "", user.Model)
	if err := gomodels.Register(app); err != nil {
		t.Fatal(err)
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
	t.Run("AddDuplicateModel", func(t *testing.T) {
		op := CreateModel{Name: "User"}
		if err := op.SetState(appState); err == nil {
			t.Errorf("expected duplicate model error")
		}
	})
	t.Run("AddModel", func(t *testing.T) {
		op := CreateModel{
			Name: "Customer",
			Fields: gomodels.Fields{
				"name": gomodels.CharField{MaxLength: 100},
			},
		}
		if err := op.SetState(appState); err != nil {
			t.Fatal(err)
		}
		model, ok := appState.models["Customer"]
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
		model := appState.models["User"]
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
		model := appState.models["User"]
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
		if _, found := appState.models["User"]; found {
			t.Errorf("model was not removed from state")
		}
	})
}

func TestModelOperations(t *testing.T) {
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{MaxLength: 100, Index: true},
		},
		gomodels.Options{},
	)
	app := gomodels.NewApp("test", "", user.Model)
	if err := gomodels.Register(app); err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	defer gomodels.Stop()
	db := gomodels.Databases()["default"]
	tx, err := db.BeginTx()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("AddModel", func(t *testing.T) {
		testAddModelOperation(t, tx, appState)
	})
	t.Run("DeleteModel", func(t *testing.T) {
		testDeleteModelOperation(t, tx, appState)
	})
	t.Run("AddIndex", func(t *testing.T) {
		testAddIndexOperation(t, tx, appState)
	})
	t.Run("RemoveIndex", func(t *testing.T) {
		testRemoveIndexOperation(t, tx, appState)
	})
}

func testAddModelOperation(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := CreateModel{Name: "Customer"}
	model := gomodels.New(
		"Customer", gomodels.Fields{}, gomodels.Options{},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User":     prevState.models["User"],
			"Customer": model,
		},
	}
	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.CreateTable = fmt.Errorf("db error")
		if err := op.Run(tx, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(tx, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("CreateTable") != 1 {
			t.Errorf("expected engine CreateTable to be called")
		}
	})
	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropTable = fmt.Errorf("db error")
		if err := op.Backwards(tx, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(tx, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropTable") != 1 {
			t.Errorf("expected engine DropTable to be called")
		}
	})
}

func testDeleteModelOperation(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := DeleteModel{Name: "User"}
	state := &AppState{
		app:    prevState.app,
		models: map[string]*gomodels.Model{},
	}
	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropTable = fmt.Errorf("db error")
		if err := op.Run(tx, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(tx, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropTable") != 1 {
			t.Errorf("expected engine DropTable to be called")
		}
	})
	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.CreateTable = fmt.Errorf("db error")
		if err := op.Backwards(tx, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(tx, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("CreateTable") != 1 {
			t.Errorf("expected engine CreateTable to be called")
		}
	})
}

func testAddIndexOperation(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := AddIndex{
		Model:  "User",
		Name:   "test_index",
		Fields: []string{"email"},
	}
	indexes := prevState.models["User"].Indexes()
	indexes["test_index"] = []string{"email"}
	model := gomodels.New(
		"User",
		prevState.models["User"].Fields(),
		gomodels.Options{Indexes: indexes},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddIndex = fmt.Errorf("db error")
		if err := op.Run(tx, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(tx, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddIndex") != 1 {
			t.Errorf("expected engine AddIndex to be called")
		}
	})
	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropIndex = fmt.Errorf("db error")
		if err := op.Backwards(tx, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(tx, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropIndex") != 1 {
			t.Errorf("expected engine DropIndex to be called")
		}
	})
}

func testRemoveIndexOperation(
	t *testing.T,
	tx *gomodels.Transaction,
	prevState *AppState,
) {
	mockedEngine := tx.Engine.(gomodels.MockedEngine)
	op := RemoveIndex{
		Model: "User",
		Name:  "test_index",
	}
	indexes := prevState.models["User"].Indexes()
	delete(indexes, "test_user_email_auto_idx")
	model := gomodels.New(
		"User",
		prevState.models["User"].Fields(),
		gomodels.Options{Indexes: indexes},
	).Model
	state := &AppState{
		app: prevState.app,
		models: map[string]*gomodels.Model{
			"User": model,
		},
	}
	t.Run("RunError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DropIndex = fmt.Errorf("db error")
		if err := op.Run(tx, state, prevState); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("RunSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Run(tx, state, prevState); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DropIndex") != 1 {
			t.Errorf("expected engine DropIndex to be called")
		}
	})
	t.Run("BackwardsError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.AddIndex = fmt.Errorf("db error")
		if err := op.Backwards(tx, prevState, state); err == nil {
			t.Errorf("expected db error")
		}
	})
	t.Run("BackwardsSuccess", func(t *testing.T) {
		mockedEngine.Reset()
		if err := op.Backwards(tx, prevState, state); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("AddIndex") != 1 {
			t.Errorf("expected engine AddIndex to be called")
		}
	})
}
