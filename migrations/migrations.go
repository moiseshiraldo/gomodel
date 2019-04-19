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
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	fp := filepath.Join(m.Path, m.Name+".json")
	if err := ioutil.WriteFile(fp, data, 0644); err != nil {
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	return nil
}

func (m *Node) Load() error {
	fp := filepath.Join(m.Path, m.Name+".json")
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	if err := json.Unmarshal(data, m); err != nil {
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	return nil
}

func Make(appName string) ([]*Node, error) {
	migrations := []*Node{}
	app, ok := gomodels.Registry[appName]
	if !ok {
		return migrations, fmt.Errorf(
			"migrations: %s: app doesn't exist", appName,
		)
	}
	if err := loadHistory(); err != nil {
		return migrations, fmt.Errorf("migrations: %v", err)
	}
	state := history[appName]
	node := &Node{
		App:          appName,
		Path:         filepath.Join(app.FullPath(), MigrationsDir),
		Dependencies: [][]string{},
		Operations:   OperationList{},
	}
	node.Name = state.nextMigrationName()
	if len(state.migrations) > 0 {
		lastNode := state.migrations[len(state.migrations)-1]
		node.Dependencies = append(
			node.Dependencies, []string{appName, lastNode.Name},
		)
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
	for _, m := range migrations {
		if err := m.Save(); err != nil {
			return migrations, fmt.Errorf("migrations: %s: %v", appName, err)
		}
	}
	return migrations, nil
}

func getModelChanges(model *gomodels.Model) OperationList {
	operations := OperationList{}
	state := history[model.App().Name()]
	modelState, ok := state.Models[model.Name()]
	if !ok {
		operation := CreateModel{Model: model.Name(), Fields: model.Fields()}
		operations = append(operations, operation)
	} else {
		newFields := gomodels.Fields{}
		removedFields := []string{}
		for name := range modelState.Fields() {
			if _, ok := model.Fields()[name]; !ok {
				removedFields = append(removedFields, name)
			}
		}
		if len(removedFields) > 0 {
			operation := RemoveFields{
				Model:  model.Name(),
				Fields: removedFields,
			}
			operations = append(operations, operation)
		}
		for name, field := range model.Fields() {
			if _, ok := modelState.Fields()[name]; !ok {
				newFields[name] = field
			}
		}
		if len(newFields) > 0 {
			operation := AddFields{Model: model.Name(), Fields: newFields}
			operations = append(operations, operation)
		}
	}
	return operations
}
