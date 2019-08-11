package migrations

import (
	"encoding/json"
	"fmt"
	_ "github.com/gwenn/gosqlite"
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
	op.back = true
	return nil
}

var mockedNodeFile = []byte(`{
  "App": "test",
  "Dependencies": [],
  "Operations": [{"MockedOperation": {}}]
}`)

func testNodeLoadNoPath(t *testing.T) {
	node := &Node{App: "test", Name: "initial", number: 1}
	if err := node.Load(); err == nil {
		t.Errorf("Expected error")
	}
}

func testNodeSaveNoPath(t *testing.T) {
	node := &Node{App: "test", Name: "initial", number: 1}
	if err := node.Save(); err == nil {
		t.Errorf("Expected error")
	}
}

func testNodeLoad(t *testing.T) {
	node := &Node{Name: "initial", number: 1, Path: tmpDir}
	if err := node.Load(); err != nil {
		t.Errorf("%s", err)
	}
	if node.App != "test" || len(node.Operations) != 1 {
		t.Errorf("node missing information")
	}
}

func testNodeSave(t *testing.T) {
	node := &Node{
		App:        "test",
		Name:       "test_migration",
		number:     2,
		Path:       tmpDir,
		Operations: OperationList{&mockedOperation{}},
	}
	if err := node.Save(); err != nil {
		t.Errorf("%s", err)
	}
	fp := filepath.Join(
		build.Default.GOPATH, "src", tmpDir, "0002_test_migration.json",
	)
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		t.Errorf("%s", err)
	}
	n := &Node{}
	if err := json.Unmarshal(data, n); err != nil {
		t.Errorf("%s", err)
	}
	if n.App != "test" || len(n.Operations) != 1 {
		t.Errorf("file missing information")
	}
}

func testNodeRunOpError(t *testing.T, db gomodels.Database) {
	op := &mockedOperation{runErr: true}
	node := &Node{
		App:        "test",
		Name:       "initial",
		number:     1,
		Operations: OperationList{op},
	}
	err := node.Run(db)
	if _, ok := err.(*OperationRunError); !ok {
		t.Errorf("Expected OperationRunError, got %T", err)
	}
}

func testNodeRunMigrationDbError(t *testing.T, db gomodels.Database) {
	op := &mockedOperation{}
	node := &Node{
		App:        "test",
		Name:       "initial",
		number:     1,
		Operations: OperationList{op},
	}
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	mockedEngine.Results.SaveMigration = fmt.Errorf("db error")
	err := node.Run(db)
	if _, ok := err.(*gomodels.DatabaseError); !ok {
		t.Errorf("Expected gomodels.DatabaseError, got %T", err)
	}
}

func testNodeTxCommitError(t *testing.T, db gomodels.Database) {
	op := &mockedOperation{}
	node := &Node{
		App:        "test",
		Name:       "initial",
		number:     1,
		Operations: OperationList{op},
	}
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	mockedEngine.Results.CommitTx = fmt.Errorf("db error")
	err := node.Run(db)
	if _, ok := err.(*gomodels.DatabaseError); !ok {
		t.Errorf("Expected gomodels.DatabaseError, got %T", err)
	}
}

func testNodeTxRollbackError(t *testing.T, db gomodels.Database) {
	op := &mockedOperation{runErr: true}
	node := &Node{
		App:        "test",
		Name:       "initial",
		number:     1,
		Operations: OperationList{op},
	}
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	mockedEngine.Results.RollbackTx = fmt.Errorf("db error")
	err := node.Run(db)
	if _, ok := err.(*gomodels.DatabaseError); !ok {
		t.Errorf("Expected gomodels.DatabaseError, got %T", err)
	}
}

func testNodeRun(t *testing.T, db gomodels.Database) {
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	op := &mockedOperation{}
	node := &Node{
		App:        "test",
		Name:       "initial",
		number:     1,
		Operations: OperationList{op},
	}
	if err := node.Run(db); err != nil {
		t.Errorf("%s", err)
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
}

func testNodeRunDependencies(t *testing.T, db gomodels.Database) {
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	op := &mockedOperation{}
	firstNode := &Node{
		App:        "test",
		Name:       "initial",
		number:     1,
		Operations: OperationList{op},
	}
	secondNode := &Node{
		App:          "test",
		Name:         "test_migration",
		number:       2,
		Dependencies: [][]string{{"test", "0001_initial"}},
	}
	if err := gomodels.Register(gomodels.NewApp("test", "")); err != nil {
		panic(err)
	}
	appState := &AppState{
		app:        gomodels.Registry()["test"],
		migrations: []*Node{firstNode, secondNode},
	}
	history["test"] = appState
	defer clearHistory()
	defer gomodels.ClearRegistry()
	if err := secondNode.Run(db); err != nil {
		t.Errorf("%s", err)
	}
	if !op.run {
		t.Errorf("node did not run operation")
	}
	if !firstNode.applied {
		t.Errorf("node was not applied")
	}
	if mockedEngine.Calls("SaveMigration") != 2 {
		t.Errorf("migrations were not saved on db")
	}
	args := mockedEngine.Args.SaveMigration
	if args.App != "test" || args.Number != 2 || args.Name != "test_migration" {
		t.Errorf("SaveMigration called with wrong arguments")
	}
}

func testNodeBackward(t *testing.T, db gomodels.Database) {
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	op := &mockedOperation{}
	node := &Node{
		App:        "test",
		Name:       "initial",
		number:     1,
		Operations: OperationList{op},
		applied:    true,
	}
	if err := node.Backwards(db); err != nil {
		t.Errorf("%s", err)
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
}

func testNodeBackwardDependencies(t *testing.T, db gomodels.Database) {
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	op := &mockedOperation{}
	firstNode := &Node{
		App:     "test",
		Name:    "initial",
		number:  1,
		applied: true,
	}
	secondNode := &Node{
		App:          "test",
		Name:         "test_migrations",
		number:       2,
		Dependencies: [][]string{{"test", "0001_initial"}},
		Operations:   OperationList{op},
		applied:      true,
	}
	if err := gomodels.Register(gomodels.NewApp("test", "")); err != nil {
		panic(err)
	}
	appState := &AppState{
		app:        gomodels.Registry()["test"],
		migrations: []*Node{firstNode, secondNode},
	}
	history["test"] = appState
	defer clearHistory()
	defer gomodels.ClearRegistry()
	if err := firstNode.Backwards(db); err != nil {
		t.Errorf("%s", err)
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
}

func clearTmp() {
	dir := filepath.Join(build.Default.GOPATH, "src", tmpDir)
	os.RemoveAll(dir)
}

func TestNodeStorage(t *testing.T) {
	dir := filepath.Join(build.Default.GOPATH, "src", tmpDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	defer clearTmp()
	filepath := filepath.Join(dir, "0001_initial.json")
	if err := ioutil.WriteFile(filepath, mockedNodeFile, 0644); err != nil {
		panic(err)
	}
	RegisterOperation("MockedOperation", &mockedOperation{})
	t.Run("LoadNoPath", testNodeLoadNoPath)
	t.Run("Load", testNodeLoad)
	t.Run("SaveNoPath", testNodeSaveNoPath)
	t.Run("Save", testNodeSave)
}

func TestNode(t *testing.T) {
	matrix := map[string]func(t *testing.T, db gomodels.Database){
		"RunOpError":           testNodeRunOpError,
		"RunMigrationDbError":  testNodeRunMigrationDbError,
		"RunTxCommitError":     testNodeTxCommitError,
		"RunTxRollbackError":   testNodeTxRollbackError,
		"Run":                  testNodeRun,
		"RunDependencies":      testNodeRunDependencies,
		"Backward":             testNodeBackward,
		"BackwardDependencies": testNodeBackwardDependencies,
	}
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		panic(err)
	}
	db := gomodels.Databases()["default"]
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	defer gomodels.Stop()
	for name, f := range matrix {
		t.Run(name, func(t *testing.T) { f(t, db) })
		mockedEngine.Reset()
	}
}
