package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

const MigrationsDir = "migrations"

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

func (m *Node) Save() error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	fp := filepath.Join(m.Path, m.Name+".json")
	if err := ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
	}
	return nil
}

func (m *Node) Load() error {
	fp := filepath.Join(m.Path, m.Name+".json")
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, m); err != nil {
		return err
	}
	return nil
}

type MakeOptions struct {
	Name      string
	Empty     bool
	OmitWrite bool
}

func Make(appName string, options MakeOptions) ([]*Node, error) {
	migrations := []*Node{}
	app, ok := gomodels.Registry[appName]
	if !ok {
		return migrations, &AppNotFoundError{Name: appName}
	}
	if err := loadHistory(); err != nil {
		return migrations, err
	}
	state := history[appName]
	node := &Node{
		App:          appName,
		Dependencies: [][]string{},
		Operations:   OperationList{},
	}
	if app.Path() != "" {
		node.Path = filepath.Join(app.FullPath(), MigrationsDir)
	}
	node.Name = state.nextMigrationFilename(options.Name)
	if len(state.migrations) > 0 {
		lastNode := state.migrations[len(state.migrations)-1]
		node.Dependencies = append(
			node.Dependencies, []string{appName, lastNode.Name},
		)
	}
	if options.Empty {
		migrations = append(migrations, node)
		return migrations, nil
	}
	for name := range state.Models {
		if _, ok := app.Models()[name]; !ok {
			node.Operations = append(node.Operations, DeleteModel{Model: name})
		}
	}
	for _, model := range app.Models() {
		node.Operations = append(node.Operations, getModelChanges(model)...)
	}
	if len(node.Operations) > 0 {
		migrations = append(migrations, node)
	}
	if options.OmitWrite {
		return migrations, nil
	} else if app.Path() == "" {
		return migrations, &PathError{
			gomodels.ErrorTrace{App: app, Err: fmt.Errorf("no path")},
		}
	}
	for _, m := range migrations {
		if err := m.Save(); err != nil {
			err = &SaveError{m.Name, gomodels.ErrorTrace{App: app, Err: err}}
			return migrations, err
		}
	}
	return migrations, nil
}
