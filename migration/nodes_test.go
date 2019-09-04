package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"os"
	"testing"
	"time"
)

// invalidOperationType mocks non-JSON-serializable operation
type invalidOperationType struct {
	InvalidField func() string
	mockedOperation
}

type fileMocker struct {
	name string
	dir  bool
}

func (f fileMocker) Name() string {
	return f.name
}

func (f fileMocker) Size() int64 {
	return 0
}

func (f fileMocker) Mode() os.FileMode {
	return 0644
}

func (f fileMocker) ModTime() time.Time {
	return time.Now()
}

func (f fileMocker) IsDir() bool {
	return f.dir
}

func (f fileMocker) Sys() interface{} {
	return nil
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

// TestReadAppNodes tests the read app nodes function
func TestReadAppNodes(t *testing.T) {
	origReadDir := readDir
	defer func() { readDir = origReadDir }()

	t.Run("Error", func(t *testing.T) {
		readDir = func(name string) ([]os.FileInfo, error) {
			return nil, fmt.Errorf("dir not found")
		}
		if _, err := readAppNodes("/app/migrations"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		readDir = func(name string) ([]os.FileInfo, error) {
			files := []os.FileInfo{
				fileMocker{name: "0001_initial.json", dir: false},
				fileMocker{name: "subdir", dir: true},
			}
			return files, nil
		}
		nodes, err := readAppNodes("/app/migrations")
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes) != 1 {
			t.Fatalf("expected 1 node, got %d", len(nodes))
		}
	})
}

// TestNode tests node Run/Backwards functions
func TestNode(t *testing.T) {
	// DB setup
	err := gomodel.Start(map[string]gomodel.Database{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	db := gomodel.Databases()["default"]
	defer gomodel.Stop()
	t.Run("Run", func(t *testing.T) { testNodeRun(t, db) })
	t.Run("Fake", func(t *testing.T) { testNodeFake(t, db) })
	t.Run("Backwards", func(t *testing.T) { testNodeBackwards(t, db) })
	t.Run("FakeBackwards", func(t *testing.T) { testNodeFakeBackwards(t, db) })
}

func testNodeRun(t *testing.T, db gomodel.Database) {
	mockedEngine := db.Engine.(gomodel.MockedEngine)
	op := &mockedOperation{}
	setup := func() *Node {
		op.reset()
		mockedEngine.Reset()
		mockedEngine.Results.TxSupport = true
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
		mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("MigrationDbRollbackError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("BeginTxError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.BeginTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("TxCommitError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.CommitTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("TxRollbackError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Run(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
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
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Errorf("migration was not saved on db")
		}
		args := mockedEngine.Args.InsertRow.Values
		if args["app"].(string) != "test" || args["number"].(int) != 1 {
			t.Errorf("InsertRow called with wrong arguments")
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
		gomodel.Register(gomodel.NewApp("test", ""))
		appState := &AppState{
			app:        gomodel.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodel.ClearRegistry()
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
		gomodel.Register(gomodel.NewApp("test", ""))
		appState := &AppState{
			app:        gomodel.Registry()["test"],
			migrations: []*Node{node, secondNode, thirdNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodel.ClearRegistry()
		if err := thirdNode.Run(db); err != nil {
			t.Fatal(err)
		}
		if !op.run {
			t.Errorf("node did not run operation")
		}
		if !node.applied {
			t.Errorf("node was not applied")
		}
		if mockedEngine.Calls("InsertRow") != 3 {
			t.Errorf("migrations were not saved on db")
		}
		args := mockedEngine.Args.InsertRow.Values
		if args["app"].(string) != "test" || args["number"].(int) != 3 {
			t.Errorf("InsertRow called with wrong arguments")
		}
	})
}

func testNodeFake(t *testing.T, db gomodel.Database) {
	mockedEngine := db.Engine.(gomodel.MockedEngine)
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

	t.Run("MigrationDbError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
		err := node.Fake(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("Skip", func(t *testing.T) {
		node := setup()
		node.applied = true
		if err := node.Fake(db); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		node := setup()
		if err := node.Fake(db); err != nil {
			t.Fatal(err)
		}
		if !node.applied {
			t.Errorf("node was not applied")
		}
		if op.run {
			t.Errorf("node did run operation")
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Errorf("migration was not saved on db")
		}
		args := mockedEngine.Args.InsertRow.Values
		if args["app"].(string) != "test" || args["number"].(int) != 1 {
			t.Errorf("InsertRow called with wrong arguments")
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
		gomodel.Register(gomodel.NewApp("test", ""))
		appState := &AppState{
			app:        gomodel.Registry()["test"],
			migrations: []*Node{node, secondNode, thirdNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodel.ClearRegistry()
		if err := thirdNode.Fake(db); err != nil {
			t.Fatal(err)
		}
		if op.run {
			t.Errorf("node did run operation")
		}
		if !node.applied {
			t.Errorf("node was not applied")
		}
		if mockedEngine.Calls("InsertRow") != 3 {
			t.Errorf("migrations were not saved on db")
		}
		args := mockedEngine.Args.InsertRow.Values
		if args["app"].(string) != "test" || args["number"].(int) != 3 {
			t.Errorf("InsertRow called with wrong arguments")
		}
	})
}

func testNodeBackwards(t *testing.T, db gomodel.Database) {
	mockedEngine := db.Engine.(gomodel.MockedEngine)
	op := &mockedOperation{}
	setup := func() *Node {
		op.reset()
		mockedEngine.Reset()
		mockedEngine.Results.TxSupport = true
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
		mockedEngine.Results.DeleteRows.Err = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("MigrationDbRollbackError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.DeleteRows.Err = fmt.Errorf("db error")
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("BeginTxError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.BeginTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("TxCommitError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.CommitTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("TxRollbackError", func(t *testing.T) {
		node := setup()
		op.RunErr = true
		mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
		err := node.Backwards(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
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
		if mockedEngine.Calls("DeleteRows") != 1 {
			t.Errorf("migration was not deleted from db")
		}
		args := mockedEngine.Args.DeleteRows.Conditioner.Conditions()
		if args["app"].(string) != "test" || args["number"].(int) != 1 {
			t.Errorf("DeleteRows called with wrong arguments")
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
		gomodel.Register(gomodel.NewApp("test", ""))
		appState := &AppState{
			app:        gomodel.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodel.ClearRegistry()
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
		gomodel.Register(gomodel.NewApp("test", ""))
		appState := &AppState{
			app:        gomodel.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodel.ClearRegistry()
		if err := node.Backwards(db); err != nil {
			t.Fatal(err)
		}
		if !op.back {
			t.Errorf("node did not run backward operation")
		}
		if secondNode.applied {
			t.Errorf("node is still applied")
		}
		if mockedEngine.Calls("DeleteRows") != 2 {
			t.Errorf("migrations were not deleted from db")
		}
		args := mockedEngine.Args.DeleteRows.Conditioner.Conditions()
		if args["app"].(string) != "test" || args["number"].(int) != 1 {
			t.Errorf("DeleteRows called with wrong arguments")
		}
	})
}

func testNodeFakeBackwards(t *testing.T, db gomodel.Database) {
	mockedEngine := db.Engine.(gomodel.MockedEngine)
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

	t.Run("MigrationDbError", func(t *testing.T) {
		node := setup()
		mockedEngine.Results.DeleteRows.Err = fmt.Errorf("db error")
		err := node.FakeBackwards(db)
		if _, ok := err.(*gomodel.DatabaseError); !ok {
			t.Errorf("expected gomodel.DatabaseError, got %T", err)
		}
	})

	t.Run("Skip", func(t *testing.T) {
		node := setup()
		node.applied = false
		if err := node.FakeBackwards(db); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		node := setup()
		if err := node.FakeBackwards(db); err != nil {
			t.Fatal(err)
		}
		if op.back {
			t.Errorf("node did run backward operation")
		}
		if node.applied {
			t.Errorf("node is still applied")
		}
		if mockedEngine.Calls("DeleteRows") != 1 {
			t.Errorf("migration was not deleted from db")
		}
		args := mockedEngine.Args.DeleteRows.Conditioner.Conditions()
		if args["app"].(string) != "test" || args["number"].(int) != 1 {
			t.Errorf("DeleteRows called with wrong arguments")
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
		gomodel.Register(gomodel.NewApp("test", ""))
		appState := &AppState{
			app:        gomodel.Registry()["test"],
			migrations: []*Node{node, secondNode},
		}
		history["test"] = appState
		defer clearHistory()
		defer gomodel.ClearRegistry()
		if err := node.FakeBackwards(db); err != nil {
			t.Fatal(err)
		}
		if op.back {
			t.Errorf("node did run backward operation")
		}
		if secondNode.applied {
			t.Errorf("node is still applied")
		}
		if mockedEngine.Calls("DeleteRows") != 2 {
			t.Errorf("migrations were not deleted from db")
		}
		args := mockedEngine.Args.DeleteRows.Conditioner.Conditions()
		if args["app"].(string) != "test" || args["number"].(int) != 1 {
			t.Errorf("DeleteRows called with wrong arguments")
		}
	})
}
