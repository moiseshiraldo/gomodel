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

func (node Node) number() int {
	number, _ := strconv.Atoi(node.Name[:4])
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
				n.Name,
				&op,
				gomodels.ErrorTrace{App: gomodels.Registry[n.App], Err: err},
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return &gomodels.DatabaseError{"", gomodels.ErrorTrace{Err: err}}
	}
	return nil
}
