package migrations

import (
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

func testAppStateNoMigrations(t *testing.T, appState *AppState) {
	err := appState.Migrate("default", "")
	if _, ok := err.(*NoAppMigrationsError); !ok {
		t.Errorf("Expected NoAppMigrationsError, got %T", err)
	}
}

func testAppStateNoDatabase(t *testing.T, appState *AppState) {
	err := appState.Migrate("SlaveDB", "")
	if _, ok := err.(*gomodels.DatabaseError); !ok {
		t.Errorf("Expected gomodels.DatabaseError, got %T", err)
	}
}

func testAppSatateInvalidNodeName(t *testing.T, appState *AppState) {
	err := appState.Migrate("default", "TestName")
	if _, ok := err.(*NameError); !ok {
		t.Errorf("Expected NameError, got %T", err)
	}
}

func testAppSatateInvalidNodeNumber(t *testing.T, appState *AppState) {
	err := appState.Migrate("default", "0003_test_migration")
	if _, ok := err.(*NameError); !ok {
		t.Errorf("Expected NameError, got %T", err)
	}
}

func testAppSatateMigrate(t *testing.T, appState *AppState) {
	err := appState.Migrate("default", "")
	if _, ok := err.(*NameError); !ok {
		t.Errorf("Expected NameError, got %T", err)
	}
}

func testAppSatateMigrateFirstNode(t *testing.T, appState *AppState) {
	if err := appState.Migrate("default", "0001"); err != nil {
		t.Errorf("%s", err)
	}
	if !appState.migrations[0].applied {
		t.Errorf("First migration was not applied")
	}
	if appState.migrations[1].applied {
		t.Errorf("Second migration was applied")
	}
}

func testAppSatateMigrateAll(t *testing.T, appState *AppState) {
	if err := appState.Migrate("default", ""); err != nil {
		t.Errorf("%s", err)
	}
	if !appState.migrations[0].applied {
		t.Errorf("First migration was not applied")
	}
	if !appState.migrations[1].applied {
		t.Errorf("Second migration was not applied")
	}
}

func testAppSatateMigrateBackwardsFirst(t *testing.T, appState *AppState) {
	if err := appState.Migrate("default", "0001"); err != nil {
		t.Errorf("%s", err)
	}
	if !appState.migrations[0].applied {
		t.Errorf("First migration is not applied")
	}
	if appState.migrations[1].applied {
		t.Errorf("Second migration is still applied")
	}
}

func testAppSatateMigrateBackwardsAll(t *testing.T, appState *AppState) {
	if err := appState.Migrate("default", "0000"); err != nil {
		t.Errorf("%s", err)
	}
	if appState.migrations[0].applied {
		t.Errorf("First migration is still applied")
	}
	if appState.migrations[1].applied {
		t.Errorf("Second migration is still applied")
	}
}

func TestAppState(t *testing.T) {
	if err := gomodels.Register(gomodels.NewApp("test", "")); err != nil {
		panic(err)
	}
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "sqlite3", Name: ":memory:"},
	})
	if err != nil {
		panic(err)
	}
	if err := gomodels.Databases()["default"].PrepareMigrations(); err != nil {
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
		testAppStateNoMigrations(t, appState)
	})
	appState.migrations = []*Node{firstNode, secondNode}
	t.Run("NoDatabase", func(t *testing.T) {
		testAppStateNoDatabase(t, appState)
	})
	t.Run("InvalidNodeName", func(t *testing.T) {
		testAppSatateInvalidNodeName(t, appState)
	})
	t.Run("InvalidNodeNumber", func(t *testing.T) {
		testAppSatateInvalidNodeNumber(t, appState)
	})
	t.Run("MigrateFirstNode", func(t *testing.T) {
		testAppSatateMigrateFirstNode(t, appState)
	})
	firstNode.applied = false
	secondNode.applied = false
	t.Run("MigrateAll", func(t *testing.T) {
		testAppSatateMigrateAll(t, appState)
	})
	firstNode.applied = true
	secondNode.applied = true
	appState.lastApplied = 2
	t.Run("MigrateBackwardsFirst", func(t *testing.T) {
		testAppSatateMigrateBackwardsFirst(t, appState)
	})
	firstNode.applied = true
	secondNode.applied = true
	appState.lastApplied = 2
	t.Run("MigrateBackwardsAll", func(t *testing.T) {
		testAppSatateMigrateBackwardsAll(t, appState)
	})
}
