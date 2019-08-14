package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const tmpDir = "github.com/moiseshiraldo/gomodels/tmp/"

type mockedOperation struct {
	run    bool
	back   bool
	state  bool
	runErr bool
}

func (op *mockedOperation) reset() {
	op.run = false
	op.back = false
	op.state = false
	op.runErr = false
}

func (op mockedOperation) OpName() string {
	return "MockedOperation"
}

func (op *mockedOperation) SetState(state *AppState) error {
	op.state = true
	return nil
}

func (op *mockedOperation) Run(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	if op.runErr {
		return fmt.Errorf("run error")
	}
	op.run = true
	return nil
}

func (op *mockedOperation) Backwards(
	tx *gomodels.Transaction,
	state *AppState,
	prevState *AppState,
) error {
	if op.runErr {
		return fmt.Errorf("run error")
	}
	op.back = true
	return nil
}

func clearTmp() {
	dir := filepath.Join(build.Default.GOPATH, "src", tmpDir)
	os.RemoveAll(dir)
}

func TestNodeStorage(t *testing.T) {
	mockedNodeFile := []byte(`{
	  "App": "test",
	  "Dependencies": [],
	  "Operations": [{"MockedOperation": {}}]
	}`)
	dir := filepath.Join(build.Default.GOPATH, "src", tmpDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	defer clearTmp()
	fp := filepath.Join(dir, "0001_initial.json")
	if err := ioutil.WriteFile(fp, mockedNodeFile, 0644); err != nil {
		t.Fatal(err)
	}
	RegisterOperation("MockedOperation", &mockedOperation{})
	t.Run("LoadNoPath", func(t *testing.T) {
		node := &Node{App: "test", Name: "initial", number: 1}
		if err := node.Load(); err == nil {
			t.Errorf("Expected error")
		}
	})
	t.Run("LoadSuccess", func(t *testing.T) {
		node := &Node{Name: "initial", number: 1, Path: tmpDir}
		if err := node.Load(); err != nil {
			t.Fatal(err)
		}
		if node.App != "test" || len(node.Operations) != 1 {
			t.Errorf("node missing information")
		}
	})
	t.Run("SaveNoPath", func(t *testing.T) {
		node := &Node{App: "test", Name: "initial", number: 1}
		if err := node.Save(); err == nil {
			t.Errorf("Expected error")
		}
	})
	t.Run("SaveSuccess", func(t *testing.T) {
		node := &Node{
			App:        "test",
			Name:       "test_migration",
			number:     2,
			Path:       tmpDir,
			Operations: OperationList{&mockedOperation{}},
		}
		if err := node.Save(); err != nil {
			t.Fatal(err)
		}
		fp := filepath.Join(
			build.Default.GOPATH, "src", tmpDir, "0002_test_migration.json",
		)
		data, err := ioutil.ReadFile(fp)
		if err != nil {
			t.Fatal(err)
		}
		n := &Node{}
		if err := json.Unmarshal(data, n); err != nil {
			t.Fatal(err)
		}
		if n.App != "test" || len(n.Operations) != 1 {
			t.Errorf("file missing information")
		}
	})
}

func TestNode(t *testing.T) {
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
	t.Run("OperationError", func(t *testing.T) {
		node := setup()
		op.runErr = true
		err := node.Run(db)
		if _, ok := err.(*OperationRunError); !ok {
			t.Errorf("Expected OperationRunError, got %T", err)
		}
	})
	t.Run("MigrationDbError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.SaveMigration = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
		}
	})
	t.Run("TxCommitError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.CommitTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
		}
	})
	t.Run("TxRollbackError", func(t *testing.T) {
		node := setup()
		op.runErr = true
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
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
	t.Run("Dependencies", func(t *testing.T) {
		node := setup()
		secondNode := &Node{
			App:          "test",
			Name:         "test_migration",
			number:       2,
			Dependencies: [][]string{{"test", "0001_initial"}},
		}
		if err := gomodels.Register(gomodels.NewApp("test", "")); err != nil {
			t.Fatal(err)
		}
		appState := &AppState{
			app:        gomodels.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodels.ClearRegistry()
		if err := secondNode.Run(db); err != nil {
			t.Fatal(err)
		}
		if !op.run {
			t.Errorf("node did not run operation")
		}
		if !node.applied {
			t.Errorf("node was not applied")
		}
		if mockedEngine.Calls("SaveMigration") != 2 {
			t.Errorf("migrations were not saved on db")
		}
		args := mockedEngine.Args.SaveMigration
		if args.App != "test" || args.Number != 2 {
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
	t.Run("OperationError", func(t *testing.T) {
		node := setup()
		op.runErr = true
		err := node.Backwards(db)
		if _, ok := err.(*OperationRunError); !ok {
			t.Errorf("Expected OperationRunError, got %T", err)
		}
	})
	t.Run("MigrationDbError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.DeleteMigration = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
		}
	})
	t.Run("TxCommitError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.CommitTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
		}
	})
	t.Run("TxRollbackError", func(t *testing.T) {
		node := setup()
		op.runErr = true
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
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
		op.reset()
		mockedEngine.Reset()
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
		if err := gomodels.Register(gomodels.NewApp("test", "")); err != nil {
			t.Fatal(err)
		}
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
