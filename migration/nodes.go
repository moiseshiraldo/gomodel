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

var writeNode = func(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}
var readNode = ioutil.ReadFile
var readDir = ioutil.ReadDir

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

type Node struct {
	App          string
	Path         string `json:"-"`
	Name         string `json:"-"`
	number       int    `json:"-"`
	processed    bool   `json:"-"`
	applied      bool   `json:"-"`
	Dependencies [][]string
	Operations   OperationList
}

func (n Node) fullname() string {
	return fmt.Sprintf("%04d_%s", n.number, n.Name)
}

func (n Node) filename() string {
	return fmt.Sprintf("%s.json", n.fullname())
}

func (n Node) Save() error {
	if n.Path == "" {
		return fmt.Errorf("no path")
	}
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return &SaveError{ErrorTrace{Node: &n, Err: err}}
	}
	fp := filepath.Join(n.Path, n.filename())
	if !filepath.IsAbs(fp) {
		fp = filepath.Join(build.Default.GOPATH, "src", fp)
	}
	if err := writeNode(fp, data); err != nil {
		return &SaveError{ErrorTrace{Node: &n, Err: err}}
	}
	return nil
}

func (n *Node) Load() error {
	if n.Path == "" {
		return fmt.Errorf("no path")
	}
	fp := filepath.Join(n.Path, n.filename())
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

func (n *Node) Run(db gomodel.Database) error {
	if n.applied {
		return nil
	}
	if err := n.runDependencies(db); err != nil {
		return err
	}
	if err := n.runOperations(db); err != nil {
		return err
	}
	n.applied = true
	return nil
}

func (n Node) runDependencies(db gomodel.Database) error {
	for _, dep := range n.Dependencies {
		app, name := dep[0], dep[1]
		number, _ := strconv.Atoi(name[:4])
		depNode := history[app].migrations[number-1]
		if !depNode.applied {
			if err := depNode.Run(db); err != nil {
				return err
			}
		}
	}
	return nil
}

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
	manager := Migration.Objects
	if txSupport {
		manager = Migration.Objects.WithTx(tx)
	}
	values := gomodel.Values{"app": n.App, "number": n.number, "name": n.Name}
	if _, err := manager.Create(values); err != nil {
		if txSupport {
			txErr := tx.Rollback()
			if txErr != nil {
				err = txErr
			}
		}
		return &gomodel.DatabaseError{db.Id(), gomodel.ErrorTrace{Err: err}}
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

func (n *Node) Backwards(db gomodel.Database) error {
	if !n.applied {
		return nil
	}
	if err := n.backwardDependencies(db); err != nil {
		return err
	}
	if err := n.backwardOperations(db); err != nil {
		return err
	}
	n.applied = false
	return nil
}

func (n Node) backwardDependencies(db gomodel.Database) error {
	for _, state := range history {
		for _, node := range state.migrations {
			for _, dep := range node.Dependencies {
				if dep[0] == n.App && dep[1] == n.fullname() {
					if err := node.Backwards(db); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

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
	for i, op := range n.Operations {
		states[i+1] = loadPreviousState(n)[n.App]
		op.SetState(states[i+1])
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
	manager := Migration.Objects
	if txSupport {
		manager = Migration.Objects.WithTx(tx)
	}
	_, err := manager.Filter(
		gomodel.Q{"app": n.App, "number": n.number},
	).Delete()
	if err != nil {
		if txSupport {
			if txErr := tx.Rollback(); txErr != nil {
				err = txErr
			}
		}
		return &gomodel.DatabaseError{db.Id(), gomodel.ErrorTrace{Err: err}}
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

func (n *Node) setState(stash map[string]map[string]bool) error {
	if n.processed {
		return nil
	}
	stash[n.App][n.fullname()] = true
	for _, dep := range n.Dependencies {
		app, depName := dep[0], dep[1]
		invalidDep := &InvalidDependencyError{
			ErrorTrace{Node: n, Err: fmt.Errorf("invalid dependency")},
		}
		if !mNameRe.MatchString(depName) {
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
		if depNode == nil || depNode.fullname() != depName {
			return invalidDep
		}
		if stash[app][depName] {
			return &CircularDependencyError{
				ErrorTrace{Node: n, Err: fmt.Errorf("invalid dependency")},
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
	delete(stash[n.App], n.Name)
	return nil
}

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
