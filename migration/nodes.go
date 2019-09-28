package migration

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"go/build"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

// writeNode holds the function to write migration nodes to a file.
var writeNode = func(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

// readNode holds the function to read migration nodes from a file.
var readNode = ioutil.ReadFile

// readDir holds the function the read app migration folders.
var readDir = ioutil.ReadDir

// readAppNodes holds a function that reads the folder given by path and
// returns all the file names inside.
var readAppNodes = func(path string) ([]string, error) {
	files, err := readDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			names = append(names, file.Name())
		}
	}
	return names, nil
}

// A Node represents a point in the application changes graph.
type Node struct {
	// App is the application name.
	App string
	// Dependencies is the list of dependencies for this node. Each dependency
	// is a list of two elements, application name and migration name
	// (e.g. ["users", "0001_initial"])
	Dependencies [][]string
	// Operations is the list of operations describing the changes.
	Operations OperationList
	path       string // Full path of the migrations folder.
	name       string // Migration name.
	number     int    // Node number.
	processed  bool   // True if node changes has been added to app state.
	applied    bool   // True if node has been applied to database schema.
}

// Name returns the node name (e.g. 0001_initial)
func (n Node) Name() string {
	return fmt.Sprintf("%04d_%s", n.number, n.name)
}

// filename returns the node filename (e.g. 0001_initial.json)
func (n Node) filename() string {
	return fmt.Sprintf("%s.json", n.Name())
}

// Save writes the the JSON representation of the node to the migrations folder.
func (n Node) Save() error {
	if n.path == "" {
		err := fmt.Errorf("no path")
		return &PathError{n.App, ErrorTrace{Node: &n, Err: err}}
	}
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return &SaveError{ErrorTrace{Node: &n, Err: err}}
	}
	fp := filepath.Join(n.path, n.filename())
	if !filepath.IsAbs(fp) {
		fp = filepath.Join(build.Default.GOPATH, "src", fp)
	}
	if err := writeNode(fp, data); err != nil {
		return &SaveError{ErrorTrace{Node: &n, Err: err}}
	}
	return nil
}

// Load reads the node details from the corresponding file in the migrations
// folder.
func (n *Node) Load() error {
	if n.path == "" {
		err := fmt.Errorf("no path")
		return &PathError{n.App, ErrorTrace{Node: n, Err: err}}
	}
	fp := filepath.Join(n.path, n.filename())
	if !filepath.IsAbs(fp) {
		fp = filepath.Join(build.Default.GOPATH, "src", fp)
	}
	data, err := readNode(fp)
	if err != nil {
		return &LoadError{ErrorTrace{Node: n, Err: err}}
	}
	if err := json.Unmarshal(data, n); err != nil {
		return &LoadError{ErrorTrace{Node: n, Err: err}}
	}
	return nil
}

// Run applies the node changes (and its unapplied dependencies) to the database
// schema given by db. All the queries will be run inside a transaction if
// supported by the database.
func (n *Node) Run(db gomodel.Database) error {
	if n.applied {
		return nil
	}
	if err := n.runDependencies(db, false); err != nil {
		return err
	}
	if err := n.runOperations(db); err != nil {
		return err
	}
	n.applied = true
	return nil
}

// Fake marks the node as applied without making any changes to the database
// schema.
func (n *Node) Fake(db gomodel.Database) error {
	if n.applied {
		return nil
	}
	if err := n.runDependencies(db, true); err != nil {
		return err
	}
	values := gomodel.Values{"app": n.App, "number": n.number, "name": n.name}
	if _, err := Migration.Objects.Create(values); err != nil {
		return err
	}
	n.applied = true
	return nil
}

// runDependencies run all the unapplied dependencies of the node.
func (n Node) runDependencies(db gomodel.Database, fake bool) error {
	for _, dep := range n.Dependencies {
		app, name := dep[0], dep[1]
		number, _ := strconv.Atoi(name[:4])
		depNode := history[app].migrations[number-1]
		if !depNode.applied {
			run := depNode.Run
			if fake {
				run = depNode.Fake
			}
			if err := run(db); err != nil {
				return err
			}
		}
	}
	return nil
}

// runOperations runs all the operations of the node (inside a transaction if
// supported by the database) and marks the node as applied.
func (n Node) runOperations(db gomodel.Database) error {
	engine := db.Engine
	txSupport := db.TxSupport()
	var tx *gomodel.Transaction
	if txSupport {
		var err error
		tx, err = db.BeginTx()
		if err != nil {
			return &gomodel.DatabaseError{
				db.Id(), gomodel.ErrorTrace{Err: err},
			}
		}
		engine = tx.Engine
	}
	prevState := loadPreviousState(n)[n.App]
	state := loadPreviousState(n)[n.App]
	for _, op := range n.Operations {
		op.SetState(state)
		if err := op.Run(engine, state, prevState); err != nil {
			if txSupport {
				if txErr := tx.Rollback(); txErr != nil {
					return &gomodel.DatabaseError{
						db.Id(), gomodel.ErrorTrace{Err: txErr},
					}
				}
			}
			return &OperationRunError{ErrorTrace{&n, op, err}}
		}
		op.SetState(prevState)
	}
	values := gomodel.Values{"app": n.App, "number": n.number, "name": n.name}
	var err error
	if txSupport {
		_, err = Migration.Objects.CreateOn(tx, values)
	} else {
		_, err = Migration.Objects.CreateOn(db.Id(), values)
	}
	if err != nil {
		if txSupport {
			txErr := tx.Rollback()
			if txErr != nil {
				return &gomodel.DatabaseError{
					db.Id(), gomodel.ErrorTrace{Err: txErr},
				}
			}
		}
		return err
	}
	if txSupport {
		if err := tx.Commit(); err != nil {
			return &gomodel.DatabaseError{
				db.Id(), gomodel.ErrorTrace{Err: err},
			}
		}
	}
	return nil
}

// Backwards reverses the node changes (and the applied nodes depending on it)
// on the database schema given by db. All the queries will be run inside a
// transaction if supported by the database.
func (n *Node) Backwards(db gomodel.Database) error {
	if !n.applied {
		return nil
	}
	if err := n.backwardDependencies(db, false); err != nil {
		return err
	}
	if err := n.backwardOperations(db); err != nil {
		return err
	}
	n.applied = false
	return nil
}

// FakeBackwards marks the node as unapplied without reversing any change on
// the database schema.
func (n *Node) FakeBackwards(db gomodel.Database) error {
	if !n.applied {
		return nil
	}
	if err := n.backwardDependencies(db, true); err != nil {
		return err
	}
	cond := gomodel.Q{"app": n.App, "number": n.number}
	if _, err := Migration.Objects.Filter(cond).Delete(); err != nil {
		return err
	}
	n.applied = false
	return nil
}

// backwardDependencies run backwards all the applied nodes depending on this.
func (n Node) backwardDependencies(db gomodel.Database, fake bool) error {
	for _, state := range history {
		for _, node := range state.migrations {
			for _, dep := range node.Dependencies {
				if dep[0] == n.App && dep[1] == n.Name() {
					goBack := node.Backwards
					if fake {
						goBack = node.FakeBackwards
					}
					if err := goBack(db); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// backwardOperations reverses all the operations of the node (inside a
// transaction if supported by the database) and marks the node as unapplied.
func (n Node) backwardOperations(db gomodel.Database) error {
	engine := db.Engine
	txSupport := db.TxSupport()
	var tx *gomodel.Transaction
	if txSupport {
		var err error
		tx, err = db.BeginTx()
		if err != nil {
			return &gomodel.DatabaseError{
				db.Id(), gomodel.ErrorTrace{Err: err},
			}
		}
		engine = tx.Engine
	}
	states := make([]*AppState, len(n.Operations)+1)
	states[0] = loadPreviousState(n)[n.App]
	for i := range n.Operations {
		states[i+1] = loadPreviousState(n)[n.App]
		for _, prevOp := range n.Operations[0 : i+1] {
			prevOp.SetState(states[i+1])
		}
	}
	for k := range n.Operations {
		i := len(n.Operations) - 1 - k
		op := n.Operations[i]
		err := op.Backwards(engine, states[i+1], states[i])
		if err != nil {
			if txSupport {
				if txErr := tx.Rollback(); txErr != nil {
					return &gomodel.DatabaseError{
						db.Id(), gomodel.ErrorTrace{Err: txErr},
					}
				}
			}
			return &OperationRunError{ErrorTrace{&n, op, err}}
		}
	}
	conditions := gomodel.Q{"app": n.App, "number": n.number}
	var err error
	if txSupport {
		_, err = Migration.Objects.Filter(conditions).WithTx(tx).Delete()
	} else {
		_, err = Migration.Objects.Filter(conditions).WithDB(db.Id()).Delete()
	}
	if err != nil {
		if txSupport {
			if txErr := tx.Rollback(); txErr != nil {
				return &gomodel.DatabaseError{
					db.Id(), gomodel.ErrorTrace{Err: txErr},
				}
			}
		}
		return err
	}
	if txSupport {
		if err := tx.Commit(); err != nil {
			return &gomodel.DatabaseError{
				db.Id(), gomodel.ErrorTrace{Err: err},
			}
		}
	}
	return nil
}

// setState applies the node changes to the application state in the global
// registry. The stash keeps a track of nodes being processed to detect
// circular dependencies.
func (n *Node) setState(stash map[string]map[string]bool) error {
	if n.processed {
		return nil
	}
	stash[n.App][n.Name()] = true
	for _, dep := range n.Dependencies {
		app, depName := dep[0], dep[1]
		invalidDep := &InvalidDependencyError{
			ErrorTrace{Node: n, Err: fmt.Errorf("invalid dependency")},
		}
		if !NodeNameRegex.MatchString(depName) {
			return invalidDep
		}
		if _, ok := history[app]; !ok {
			return invalidDep
		}
		depNumber, _ := strconv.Atoi(depName[:4])
		if depNumber > len(history[app].migrations) {
			return invalidDep
		}
		depNode := history[app].migrations[depNumber-1]
		if depNode == nil || depNode.Name() != depName {
			return invalidDep
		}
		if stash[app][depName] {
			return &CircularDependencyError{
				ErrorTrace{Node: n, Err: fmt.Errorf("circular dependency")},
			}
		}
		if !depNode.processed {
			if err := depNode.setState(stash); err != nil {
				return err
			}
		}
	}
	for _, op := range n.Operations {
		if err := op.SetState(history[n.App]); err != nil {
			return &OperationStateError{ErrorTrace{n, op, err}}
		}
	}
	n.processed = true
	delete(stash[n.App], n.Name())
	return nil
}

// setPreviousState applies the changes up to the previous node on the given
// registry of application states.
func (n Node) setPreviousState(prevHistory map[string]*AppState) {
	for _, dep := range n.Dependencies {
		app, depName := dep[0], dep[1]
		number, _ := strconv.Atoi(depName[:4])
		depNode := history[app].migrations[number-1]
		depNode.setPreviousState(prevHistory)
	}
	for _, op := range n.Operations {
		op.SetState(prevHistory[n.App])
	}
}
