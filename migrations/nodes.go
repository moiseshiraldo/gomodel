package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"go/build"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

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
		return err
	}
	fp := filepath.Join(n.Path, n.filename())
	if !filepath.IsAbs(fp) {
		fp = filepath.Join(build.Default.GOPATH, "src", fp)
	}
	if err := ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
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
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, n); err != nil {
		return err
	}
	return nil
}

func (n *Node) Run(db gomodels.Database) error {
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

func (n Node) runDependencies(db gomodels.Database) error {
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

func (n Node) runOperations(db gomodels.Database) error {
	tx, err := db.BeginTx()
	if err != nil {
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	for _, op := range n.Operations {
		if err := op.Run(tx, history[n.App]); err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				return &gomodels.DatabaseError{
					"", gomodels.ErrorTrace{Err: txErr},
				}
			}
			return &OperationRunError{
				ErrorTrace{Node: &n, Operation: &op, Err: err},
			}
		}
	}
	if err := tx.SaveMigration(n.App, n.number, n.Name); err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			err = txErr
		}
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	if err = tx.Commit(); err != nil {
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	return nil
}

func (n *Node) Backwards(db gomodels.Database) error {
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

func (n Node) backwardDependencies(db gomodels.Database) error {
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

func (n Node) backwardOperations(db gomodels.Database) error {
	tx, err := db.BeginTx()
	if err != nil {
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	prevState := loadPreviousState(n)
	for k := range n.Operations {
		op := n.Operations[len(n.Operations)-1-k]
		err := op.Backwards(tx, prevState[n.App])
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				return &gomodels.DatabaseError{
					"", gomodels.ErrorTrace{Err: txErr},
				}
			}
			return &OperationRunError{
				ErrorTrace{Node: &n, Operation: &op, Err: err},
			}
		}
	}
	if err := tx.DeleteMigration(n.App, n.number); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			err = txErr
		}
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	if err = tx.Commit(); err != nil {
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	return nil
}

func (n *Node) setState(stash map[string]map[string]bool) error {
	if n.processed {
		return nil
	}
	stash[n.App][n.Name] = true
	for _, dep := range n.Dependencies {
		app, depName := dep[0], dep[1]
		if !mNameRe.MatchString(depName) {
			return &InvalidDependencyError{ErrorTrace{Node: n}}
		}
		if _, ok := history[app]; !ok {
			return &InvalidDependencyError{ErrorTrace{Node: n}}
		}
		number, _ := strconv.Atoi(depName[:4])
		if number > len(history[app].migrations) {
			return &InvalidDependencyError{ErrorTrace{Node: n}}
		}
		depNode := history[app].migrations[number-1]
		if depNode == nil {
			return &InvalidDependencyError{ErrorTrace{Node: n}}
		}
		if _, found := stash[app][depName]; found {
			return &CircularDependencyError{ErrorTrace{Node: n}}
		}
		if !depNode.processed {
			if err := depNode.setState(stash); err != nil {
				return err
			}
		}
	}
	for _, op := range n.Operations {
		if err := op.SetState(history[n.App]); err != nil {
			return &OperationStateError{ErrorTrace{Node: n}}
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
