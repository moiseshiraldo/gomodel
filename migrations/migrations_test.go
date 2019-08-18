package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

func TestMake(t *testing.T) {
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{MaxLength: 100},
		},
		gomodels.Options{},
	)
	app := gomodels.NewApp("users", "", user.Model)
	if err := gomodels.Register(app); err != nil {
		t.Fatal(err)
	}
	defer gomodels.ClearRegistry()
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
				app:    gomodels.Registry()["users"],
				Models: make(map[string]*gomodels.Model),
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
				app:    gomodels.Registry()["users"],
				Models: make(map[string]*gomodels.Model),
			}
			return nil
		}
		_, err := Make("users", MakeOptions{})
		if _, ok := err.(*PathError); !ok {
			t.Errorf("expected PathError, got %T", err)
		}
	})
	t.Run("WriteError", func(t *testing.T) {
		gomodels.ClearRegistry()
		app.Path = "users/migrations"
		if err := gomodels.Register(app); err != nil {
			t.Fatal(err)
		}
		loadHistory = func() error {
			history["users"] = &AppState{
				app:    gomodels.Registry()["users"],
				Models: make(map[string]*gomodels.Model),
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
		gomodels.ClearRegistry()
		app.Path = "users/migrations"
		if err := gomodels.Register(app); err != nil {
			t.Fatal(err)
		}
		loadHistory = func() error {
			history["users"] = &AppState{
				app:    gomodels.Registry()["users"],
				Models: make(map[string]*gomodels.Model),
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

func TestRun(t *testing.T) {
	app := gomodels.NewApp("users", "")
	if err := gomodels.Register(app); err != nil {
		t.Fatal(err)
	}
	defer gomodels.ClearRegistry()
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gomodels.Stop()
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
			Name:   "initial",
			number: 1,
		}
		state = &AppState{
			app:        gomodels.Registry()["users"],
			Models:     make(map[string]*gomodels.Model),
			migrations: []*Node{node},
		}
		history["users"] = state
		return nil
	}
	t.Run("NoDatabase", func(t *testing.T) {
		err := Run(RunOptions{App: "users", Database: "missing"})
		if _, ok := err.(*gomodels.DatabaseError); !ok {
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
		loadAppliedMigrations = func(db gomodels.Database) error {
			loadAppliedCalled = true
			return fmt.Errorf("load applied migrations error")
		}
		err := Run(RunOptions{App: "users"})
		if !loadAppliedCalled {
			t.Fatal("expected load history to be called")
		}
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})
	t.Run("NoApp", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodels.Database) error { return nil }
		err := Run(RunOptions{App: "customers"})
		if _, ok := err.(*AppNotFoundError); !ok {
			t.Errorf("expected AppNotFoundError, got %T", err)
		}
	})
	t.Run("App", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodels.Database) error {
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
	t.Run("All", func(t *testing.T) {
		loadHistory = mockedLoadHistory
		loadAppliedMigrations = func(db gomodels.Database) error { return nil }
		err := Run(RunOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if state.migrations == nil || !state.migrations[0].applied {
			t.Fatal("migration was not applied")
		}
	})
}

func TestMakeAndRun(t *testing.T) {
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{MaxLength: 100},
		},
		gomodels.Options{},
	)
	app := gomodels.NewApp("users", "", user.Model)
	if err := gomodels.Register(app); err != nil {
		t.Fatal(err)
	}
	defer gomodels.ClearRegistry()
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gomodels.Stop()
	origLoadApplied := loadAppliedMigrations
	defer func() { loadAppliedMigrations = origLoadApplied }()
	loadAppliedCalled := false
	t.Run("NoDatabase", func(t *testing.T) {
		err := MakeAndRun("missing")
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})
	t.Run("LoadAppliedError", func(t *testing.T) {
		loadAppliedMigrations = func(db gomodels.Database) error {
			return fmt.Errorf("db error")
		}
		err := MakeAndRun("default")
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})
	t.Run("Success", func(t *testing.T) {
		db := gomodels.Databases()["default"]
		mockedEngine := db.Engine.(gomodels.MockedEngine)
		loadAppliedMigrations = func(db gomodels.Database) error {
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
		if mockedEngine.Calls("SaveMigration") != 1 {
			t.Fatal("migration was not applied")
		}
	})
}
