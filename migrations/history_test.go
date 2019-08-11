package migrations

import (
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

func TestAppState(t *testing.T) {
	if err := gomodels.Register(gomodels.NewApp("test", "")); err != nil {
		panic(err)
	}
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		panic(err)
	}
	defer gomodels.Stop()
	defer gomodels.ClearRegistry()
	firstNode := &Node{App: "test", Name: "initial", number: 1}
	secondNode := &Node{
		App:          "test",
		Name:         "test_migrations",
		number:       2,
		Dependencies: [][]string{{"test", "0001_initial"}},
	}
	appState := &AppState{
		app: gomodels.Registry()["test"],
	}
	history["test"] = appState
	defer clearHistory()
	t.Run("NoMigrations", func(t *testing.T) {
		err := appState.Migrate("default", "")
		if _, ok := err.(*NoAppMigrationsError); !ok {
			t.Errorf("Expected NoAppMigrationsError, got %T", err)
		}
	})
	appState.migrations = []*Node{firstNode, secondNode}
	t.Run("NoDatabase", func(t *testing.T) {
		err := appState.Migrate("SlaveDB", "")
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
		}
	})
	t.Run("InvalidNodeName", func(t *testing.T) {
		err := appState.Migrate("default", "TestName")
		if _, ok := err.(*NameError); !ok {
			t.Errorf("Expected NameError, got %T", err)
		}
	})
	t.Run("InvalidNodeNumber", func(t *testing.T) {
		err := appState.Migrate("default", "0003_test_migration")
		if _, ok := err.(*NameError); !ok {
			t.Errorf("Expected NameError, got %T", err)
		}
	})
	t.Run("MigrateFirstNode", func(t *testing.T) {
		if err := appState.Migrate("default", "0001"); err != nil {
			t.Errorf("%s", err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("First migration was not applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("Second migration was applied")
		}
	})
	t.Run("MigrateAll", func(t *testing.T) {
		firstNode.applied = false
		secondNode.applied = false
		if err := appState.Migrate("default", ""); err != nil {
			t.Errorf("%s", err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("First migration was not applied")
		}
		if !appState.migrations[1].applied {
			t.Errorf("Second migration was not applied")
		}
	})
	t.Run("MigrateBackwardsFirst", func(t *testing.T) {
		firstNode.applied = true
		secondNode.applied = true
		appState.lastApplied = 2
		if err := appState.Migrate("default", "0001"); err != nil {
			t.Errorf("%s", err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("First migration is not applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("Second migration is still applied")
		}
	})
	t.Run("MigrateBackwardsAll", func(t *testing.T) {
		firstNode.applied = true
		secondNode.applied = true
		appState.lastApplied = 2
		if err := appState.Migrate("default", "0000"); err != nil {
			t.Errorf("%s", err)
		}
		if appState.migrations[0].applied {
			t.Errorf("First migration is still applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("Second migration is still applied")
		}
	})
}
