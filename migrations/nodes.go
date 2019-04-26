package migrations

import (
	"database/sql"
	"encoding/json"
	"github.com/moiseshiraldo/gomodels"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

type Node struct {
	App          string
	Path         string `json:"-"`
	Name         string `json:"-"`
	processed    bool   `json:"-"`
	applied      bool   `json:"-"`
	Dependencies [][]string
	Operations   OperationList
}

func (n Node) number() int {
	number, _ := strconv.Atoi(n.Name[:4])
	return number
}

func (n *Node) Save() error {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return err
	}
	fp := filepath.Join(n.Path, n.Name+".json")
	if err := ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
	}
	return nil
}

func (n *Node) Load() error {
	fp := filepath.Join(n.Path, n.Name+".json")
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, n); err != nil {
		return err
	}
	return nil
}

func (n *Node) Run(db *sql.DB) error {
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

func (n *Node) runDependencies(db *sql.DB) error {
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

func (n *Node) runOperations(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	for _, op := range n.Operations {
		if err := op.Run(tx, n.App); err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				return &gomodels.DatabaseError{
					"", gomodels.ErrorTrace{Err: txErr},
				}
			}
			return &OperationRunError{
				ErrorTrace{Node: n, Operation: &op, Err: err},
			}
		}
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
