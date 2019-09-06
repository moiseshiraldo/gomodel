package migration

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"testing"
)

// TestMake tests the Make function
func TestMake(t *testing.T) {
	// Models setup
	user := gomodel.New(
		"User",
		gomodel.Fields{
			"email": gomodel.CharField{MaxLength: 100},
		},
		gomodel.Options{},
	)
	// App setup
	app := gomodel.NewApp("users", "", user.Model)
	gomodel.Register(app)
	defer gomodel.ClearRegistry()
	// Mocks loadHistory and writeNode functions
	origLoadHistory := loadHistory
	defer func() { loadHistory = origLoadHistory }()
	loadHistoryCalled := false
	origWriteNode := writeNode
	defer func() { writeNode = origWriteNode }()

	t.Run("NoApp", func(t *testing.T) {
		_, err := Make("test", MakeOptions{})
		if _, ok := err.(*AppNotFoundError); !ok {
			t.Errorf("expected AppNotFoundError, got %T", err)
		}
	})

	t.Run("LoadHistoryError", func(t *testing.T) {
		loadHistory = func() error {
			loadHistoryCalled = true
			return fmt.Errorf("load history error")
		}
		_, err := Make("users", MakeOptions{})
		if !loadHistoryCalled {
			t.Fatal("expected load history to be called")
		}
		if err == nil || err.Error() != "load history error" {
			t.Errorf("expected load history error, got %T", err)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		loadHistory = func() error {
			history["users"] = &AppState{
				app:    gomodel.Registry()["users"],
				Models: make(map[string]*gomodel.Model),
			}
			return nil
		}
		state, err := Make("users", MakeOptions{Empty: true})
		if err != nil {
			t.Fatal(err)
		}
		if state == nil {
			t.Fatal("expected app state to be returned")
		}
		if len(state.migrations) != 1 {
			t.Errorf(
				"app state has %d migrations, expected one",
				len(state.migrations),
			)
		}
	})

	t.Run("NoPath", func(t *testing.T) {
		loadHistory = func() error {
			history["users"] = &AppState{
				app:    gomodel.Registry()["users"],
				Models: make(map[string]*gomodel.Model),
			}
			return nil
		}
		_, err := Make("users", MakeOptions{})
		if _, ok := err.(*PathError); !ok {
			t.Errorf("expected PathError, got %T", err)
		}
	})

	t.Run("WriteError", func(t *testing.T) {
		gomodel.ClearRegistry()
		app.Path = "users/migrations"
		gomodel.Register(app)
		loadHistory = func() error {
			history["users"] = &AppState{
				app:    gomodel.Registry()["users"],
				Models: make(map[string]*gomodel.Model),
			}
			return nil
		}
		writeNode = func(path string, data []byte) error {
			return fmt.Errorf("write error")
		}
		_, err := Make("users", MakeOptions{})
		if _, ok := err.(*SaveError); !ok {
			t.Errorf("expected SaveError, got %T", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		gomodel.ClearRegistry()
		app.Path = "users/migrations"
		gomodel.Register(app)
		loadHistory = func() error {
			history["users"] = &AppState{
				app:    gomodel.Registry()["users"],
				Models: make(map[string]*gomodel.Model),
			}
			return nil
		}
		writeNodeCallData := make([]byte, 0)
		writeNode = func(path string, data []byte) error {
			writeNodeCallData = data
			return nil
		}
		_, err := Make("users", MakeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		n := &Node{}
		if err := json.Unmarshal(writeNodeCallData, n); err != nil {
			t.Fatal(err)
		}
		if n.App != "users" || len(n.Operations) != 1 {
			t.Error("file missing information")
		}
	})
}

// TestRun tests the Run method
func TestRun(t *testing.T) {
	// App setup
	app := gomodel.NewApp("users", "")
	gomodel.Register(app)
	defer gomodel.ClearRegistry()
	// DB setup
	err := gomodel.Start(map[string]gomodel.Database{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	db := gomodel.Databases()["default"]
	mockedEngine := db.Engine.(gomodel.MockedEngine)
	defer gomodel.Stop()
	// Mocks loadHistory and loadAppliedMigrations functions
	origLoadHistory := loadHistory
	origLoadApplied := loadAppliedMigrations
	defer func() {
		loadHistory = origLoadHistory
		loadAppliedMigrations = origLoadApplied
	}()
	loadHistoryCalled := false
	loadAppliedCalled := false
	state := &AppState{}
	mockedLoadHistory := func() error {
		node := &Node{
			App:    "users",
			name:   "initial",
			number: 1,
		}
		state = &AppState{
			app:        gomodel.Registry()["users"],
			Models:     make(map[string]*gomodel.Model),
			migrations: []*Node{node},
		}
		history["users"] = state
		return nil
	}

	t.Run("NoDatabase", func(t *testing.T) {
		err := Run(RunOptions{App: "users", Database: "missing"})
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("LoadHistoryError", func(t *testing.T) {
		loadHistory = func() error {
			loadHistoryCalled = true
			return fmt.Errorf("load history error")
		}
		err := Run(RunOptions{App: "users"})
		if !loadHistoryCalled {
			t.Fatal("expected load history to be called")
		}
		if err == nil || err.Error() != "load history error" {
			t.Errorf("expected load history error, got %T", err)
		}
	})

	t.Run("LoadAppliedError", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodel.Database) error {
			loadAppliedCalled = true
			return fmt.Errorf("load applied migrations error")
		}
		if err := Run(RunOptions{App: "users"}); err == nil {
			t.Fatal("expected error, got nil")
		}
		if !loadAppliedCalled {
			t.Fatal("expected load history to be called")
		}
	})

	t.Run("NoApp", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodel.Database) error { return nil }
		err := Run(RunOptions{App: "customers"})
		if _, ok := err.(*AppNotFoundError); !ok {
			t.Errorf("expected AppNotFoundError, got %T", err)
		}
	})

	t.Run("App", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodel.Database) error {
			loadAppliedCalled = true
			return nil
		}
		err := Run(RunOptions{App: "users"})
		if err != nil {
			t.Fatal(err)
		}
		if !loadAppliedCalled {
			t.Fatal("expected load applied migrations to be called")
		}
		if state.migrations == nil || !state.migrations[0].applied {
			t.Fatal("migration was not applied")
		}
	})

	t.Run("AppFake", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodel.Database) error {
			loadAppliedCalled = true
			return nil
		}
		err := Run(RunOptions{App: "users", Fake: true})
		if err != nil {
			t.Fatal(err)
		}
		if !loadAppliedCalled {
			t.Fatal("expected load applied migrations to be called")
		}
		if state.migrations == nil || !state.migrations[0].applied {
			t.Fatal("migration was not applied")
		}
	})

	t.Run("AppDBError", func(t *testing.T) {
		loadHistory = func() error {
			state = &AppState{
				app:        gomodel.Registry()["users"],
				Models:     make(map[string]*gomodel.Model),
				migrations: []*Node{},
			}
			history["users"] = state
			return nil
		}
		loadAppliedMigrations = func(db gomodel.Database) error { return nil }
		err := Run(RunOptions{App: "users"})
		if _, ok := err.(*NoAppMigrationsError); !ok {
			t.Fatalf("expected NoAppMigrationsError, got %T", err)
		}
	})

	t.Run("All", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodel.Database) error { return nil }
		err := Run(RunOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if state.migrations == nil || !state.migrations[0].applied {
			t.Fatal("migration was not applied")
		}
	})

	t.Run("AllFake", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodel.Database) error { return nil }
		err := Run(RunOptions{Fake: true})
		if err != nil {
			t.Fatal(err)
		}
		if state.migrations == nil || !state.migrations[0].applied {
			t.Fatal("migration was not applied")
		}
	})

	t.Run("AllDBError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodel.Database) error { return nil }
		err := Run(RunOptions{})
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Fatalf("expected gomodel.DatabaseError, got %T", err)
		}
	})
}

// TestMakeAndRun tests the MakeAndRun function
func TestMakeAndRun(t *testing.T) {
	// Models setup
	user := gomodel.New(
		"User",
		gomodel.Fields{
			"email": gomodel.CharField{MaxLength: 100},
		},
		gomodel.Options{},
	)
	// App setup
	app := gomodel.NewApp("users", "", user.Model)
	gomodel.Register(app)
	defer gomodel.ClearRegistry()
	// DB setup
	err := gomodel.Start(map[string]gomodel.Database{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	db := gomodel.Databases()["default"]
	mockedEngine := db.Engine.(gomodel.MockedEngine)
	defer gomodel.Stop()
	// Mocks loadAppliedMigrations function
	origLoadApplied := loadAppliedMigrations
	defer func() { loadAppliedMigrations = origLoadApplied }()
	loadAppliedCalled := false

	t.Run("NoDatabase", func(t *testing.T) {
		err := MakeAndRun("missing")
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("LoadAppliedError", func(t *testing.T) {
		loadAppliedMigrations = func(db gomodel.Database) error {
			return fmt.Errorf("db error")
		}
		err := MakeAndRun("default")
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("DBError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
		loadAppliedMigrations = func(db gomodel.Database) error { return nil }
		err := MakeAndRun("default")
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Fatalf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockedEngine.Reset()
		loadAppliedMigrations = func(db gomodel.Database) error {
			loadAppliedCalled = true
			return nil
		}
		err := MakeAndRun("default")
		if err != nil {
			t.Fatal(err)
		}
		if !loadAppliedCalled {
			t.Fatal("expected load applied migrations to be called")
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Fatal("migration was not applied")
		}
	})
}
