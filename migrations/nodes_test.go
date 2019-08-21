package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

// invalidOperationType mocks non-JSON-serializable operation
type invalidOperationType struct {
	InvalidField func() string
	mockedOperation
}

// TestNodeStorage tests node Load/Save methods
func TestNodeStorage(t *testing.T) {
	// Mocks file read/write functions
	mockedNodeFile := []byte(`{
	  "App": "test",
	  "Dependencies": [],
	  "Operations": [{"MockedOperation": {}}]
	}`)
	origReadNode := readNode
	origWriteNode := writeNode
	defer func() {
		readNode = origReadNode
		writeNode = origWriteNode
	}()
	writeNodeCalled := false
	// Registers mocked operation
	if _, ok := operationsRegistry["MockedOperation"]; !ok {
		operationsRegistry["MockedOperation"] = &mockedOperation{}
	}

	t.Run("LoadNoPath", func(t *testing.T) {
		node := &Node{App: "test", Name: "initial", number: 1}
		err := node.Load()
		if _, ok := err.(*LoadError); !ok {
			fmt.Errorf("expected LoadError, got %T", err)
		}
	})
	t.Run("LoadReadError", func(t *testing.T) {
		readNode = func(path string) ([]byte, error) {
			return nil, fmt.Errorf("read error")
		}
		node := &Node{
			App:    "test",
			Path:   "test/migrations/",
			Name:   "initial",
			number: 1,
		}
		err := node.Load()
		if _, ok := err.(*LoadError); !ok {
			fmt.Errorf("expected LoadError, got %T", err)
		}
	})

	t.Run("LoadSuccess", func(t *testing.T) {
		readNode = func(path string) ([]byte, error) {
			return mockedNodeFile, nil
		}
		node := &Node{
			App:    "test",
			Path:   "test/migrations/",
			Name:   "initial",
			number: 1,
		}
		if err := node.Load(); err != nil {
			t.Fatal(err)
		}
		if node.App != "test" || len(node.Operations) != 1 {
			t.Errorf("node missing information")
		}
	})

	t.Run("SaveNoPath", func(t *testing.T) {
		node := &Node{App: "test", Name: "initial", number: 1}
		err := node.Save()
		if _, ok := err.(*SaveError); !ok {
			fmt.Errorf("expected LoadError, got %T", err)
		}
	})

	t.Run("SaveUnknownOperation", func(t *testing.T) {
		node := &Node{
			App:        "test",
			Name:       "test_migration",
			number:     2,
			Path:       "test/migrations/",
			Operations: OperationList{&invalidOperationType{}},
		}
		err := node.Save()
		if _, ok := err.(*LoadError); !ok {
			fmt.Errorf("expected LoadError, got %T", err)
		}
	})

	t.Run("SaveWriteError", func(t *testing.T) {
		writeNode = func(path string, data []byte) error {
			return fmt.Errorf("write error")
		}
		node := &Node{
			App:        "test",
			Name:       "test_migration",
			number:     2,
			Path:       "test/migrations/",
			Operations: OperationList{&mockedOperation{}},
		}
		err := node.Save()
		if _, ok := err.(*LoadError); !ok {
			fmt.Errorf("expected LoadError, got %T", err)
		}
	})

	t.Run("SaveSuccess", func(t *testing.T) {
		writeNode = func(path string, data []byte) error {
			writeNodeCalled = true
			return nil
		}
		node := &Node{
			App:        "test",
			Name:       "test_migration",
			number:     2,
			Path:       "test/migrations/",
			Operations: OperationList{&mockedOperation{}},
		}
		if err := node.Save(); err != nil {
			t.Fatal(err)
		}
		if !writeNodeCalled {
			t.Error("expected writeNode to be called")
		}
	})
}

// TestNode tests node Run/Backwards functions
func TestNode(t *testing.T) {
	// DB setup
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	db := gomodels.Databases()["default"]
	defer gomodels.Stop()
	t.Run("Run", func(t *testing.T) { testNodeRun(t, db) })
	t.Run("Backwards", func(t *testing.T) { testNodeBackwards(t, db) })
}

func testNodeRun(t *testing.T, db gomodels.Database) {
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	op := &mockedOperation{}
	setup := func() *Node {
		op.reset()
		mockedEngine.Reset()
		return &Node{
			App:        "test",
			Name:       "initial",
			number:     1,
			Operations: OperationList{op},
		}
	}
	defer mockedEngine.Reset()

	t.Run("OperationError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		err := node.Run(db)
		if _, ok := err.(*OperationRunError); !ok {
			t.Errorf("expected OperationRunError, got %T", err)
		}
	})

	t.Run("MigrationDbError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.SaveMigration = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("BeginTxError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.BeginTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("TxCommitError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.CommitTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("TxRollbackError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("Skip", func(t *testing.T) {
		node := setup()
		node.applied = true
		if err := node.Run(db); err != nil {
			t.Fatal(err)
		}
		if op.run {
			t.Errorf("expected node to be skipped, but operation was run")
		}
	})

	t.Run("Success", func(t *testing.T) {
		node := setup()
		if err := node.Run(db); err != nil {
			t.Fatal(err)
		}
		if !op.run {
			t.Errorf("node did not run operation")
		}
		if !node.applied {
			t.Errorf("node was not applied")
		}
		if mockedEngine.Calls("SaveMigration") != 1 {
			t.Errorf("migration was not saved on db")
		}
		args := mockedEngine.Args.SaveMigration
		if args.App != "test" || args.Number != 1 || args.Name != "initial" {
			t.Errorf("SaveMigration called with wrong arguments")
		}
	})

	t.Run("DependencyError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		secondNode := &Node{
			App:          "test",
			Name:         "second",
			number:       2,
			Dependencies: [][]string{{"test", "0001_initial"}},
		}
		gomodels.Register(gomodels.NewApp("test", ""))
		appState := &AppState{
			app:        gomodels.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodels.ClearRegistry()
		err := secondNode.Run(db)
		if _, ok := err.(*OperationRunError); !ok {
			t.Errorf("expected OperationRunError, got %T", err)
		}
	})

	t.Run("Dependencies", func(t *testing.T) {
		node := setup()
		secondNode := &Node{
			App:          "test",
			Name:         "second",
			number:       2,
			Dependencies: [][]string{{"test", "0001_initial"}},
		}
		thirdNode := &Node{
			App:          "test",
			Name:         "third",
			number:       3,
			Dependencies: [][]string{{"test", "0002_second"}},
		}
		gomodels.Register(gomodels.NewApp("test", ""))
		appState := &AppState{
			app:        gomodels.Registry()["test"],
			migrations: []*Node{node, secondNode, thirdNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodels.ClearRegistry()
		if err := thirdNode.Run(db); err != nil {
			t.Fatal(err)
		}
		if !op.run {
			t.Errorf("node did not run operation")
		}
		if !node.applied {
			t.Errorf("node was not applied")
		}
		if mockedEngine.Calls("SaveMigration") != 3 {
			t.Errorf("migrations were not saved on db")
		}
		args := mockedEngine.Args.SaveMigration
		if args.App != "test" || args.Number != 3 {
			t.Errorf("SaveMigration called with wrong arguments")
		}
	})
}

func testNodeBackwards(t *testing.T, db gomodels.Database) {
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	op := &mockedOperation{}
	setup := func() *Node {
		op.reset()
		mockedEngine.Reset()
		return &Node{
			App:        "test",
			Name:       "initial",
			number:     1,
			Operations: OperationList{op},
			applied:    true,
		}
	}
	defer mockedEngine.Reset()

	t.Run("OperationError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		err := node.Backwards(db)
		if _, ok := err.(*OperationRunError); !ok {
			t.Errorf("expected OperationRunError, got %T", err)
		}
	})

	t.Run("MigrationDbError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.DeleteMigration = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("BeginTxError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.BeginTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("TxCommitError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.CommitTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("TxRollbackError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("Skip", func(t *testing.T) {
		node := setup()
		node.applied = false
		if err := node.Backwards(db); err != nil {
			t.Fatal(err)
		}
		if op.back {
			t.Errorf("expected node to be skipped, but operation was run")
		}
	})

	t.Run("Success", func(t *testing.T) {
		node := setup()
		if err := node.Backwards(db); err != nil {
			t.Fatal(err)
		}
		if !op.back {
			t.Errorf("node did not run backward operation")
		}
		if node.applied {
			t.Errorf("node is still applied")
		}
		if mockedEngine.Calls("DeleteMigration") != 1 {
			t.Errorf("migration was not deleted from db")
		}
		args := mockedEngine.Args.DeleteMigration
		if args.App != "test" || args.Number != 1 {
			t.Errorf("DeleteMigration called with wrong arguments")
		}
	})

	t.Run("DependencyError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		node.Operations = OperationList{}
		secondNode := &Node{
			App:          "test",
			Name:         "test_migrations",
			number:       2,
			Dependencies: [][]string{{"test", "0001_initial"}},
			Operations:   OperationList{op},
			applied:      true,
		}
		gomodels.Register(gomodels.NewApp("test", ""))
		appState := &AppState{
			app:        gomodels.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodels.ClearRegistry()
		err := node.Backwards(db)
		if _, ok := err.(*OperationRunError); !ok {
			t.Errorf("expected OperationRunError, got %T", err)
		}
	})

	t.Run("Dependencies", func(t *testing.T) {
		node := setup()
		secondNode := &Node{
			App:          "test",
			Name:         "test_migrations",
			number:       2,
			Dependencies: [][]string{{"test", "0001_initial"}},
			Operations:   OperationList{op},
			applied:      true,
		}
		gomodels.Register(gomodels.NewApp("test", ""))
		appState := &AppState{
			app:        gomodels.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodels.ClearRegistry()
		if err := node.Backwards(db); err != nil {
			t.Fatal(err)
		}
		if !op.back {
			t.Errorf("node did not run backward operation")
		}
		if secondNode.applied {
			t.Errorf("node is still applied")
		}
		if mockedEngine.Calls("DeleteMigration") != 2 {
			t.Errorf("migrations were not deleted from db")
		}
		args := mockedEngine.Args.DeleteMigration
		if args.App != "test" || args.Number != 1 {
			t.Errorf("DeleteMigration called with wrong arguments")
		}
	})
}
