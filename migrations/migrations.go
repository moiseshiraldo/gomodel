package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"path/filepath"
	"strconv"
)

const MigrationsDir = "migrations"

type MakeOptions struct {
	Name      string
	Empty     bool
	OmitWrite bool
}

func Make(appName string, options MakeOptions) ([]*Node, error) {
	migrations := []*Node{}
	app, ok := gomodels.Registry[appName]
	if !ok {
		return migrations, &AppNotFoundError{appName, ErrorTrace{}}
	}
	if err := loadHistory(); err != nil {
		return migrations, err
	}
	defer clearHistory()
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
			node.Operations = append(node.Operations, DeleteModel{Name: name})
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
		return migrations, &PathError{ErrorTrace{Err: fmt.Errorf("no path")}}
	}
	for _, m := range migrations {
		if err := m.Save(); err != nil {
			err = &SaveError{ErrorTrace{Node: m, Err: err}}
			return migrations, err
		}
	}
	return migrations, nil
}

type RunOptions struct {
	App      string
	Node     string
	Fake     bool
	Database string
}

func Run(options RunOptions) error {
	dbName := options.Database
	if dbName == "" {
		dbName = "default"
	}
	db, ok := gomodels.Databases[dbName]
	if !ok {
		return &gomodels.DatabaseError{dbName, gomodels.ErrorTrace{}}
	}
	if err := prepareDatabase(db); err != nil {
		return err
	}
	if err := loadHistory(); err != nil {
		return err
	}
	defer clearHistory()
	if err := loadAppliedMigrations(db); err != nil {
		return &gomodels.DatabaseError{dbName, gomodels.ErrorTrace{Err: err}}
	}
	if options.App != "" {
		state, ok := history[options.App]
		if !ok {
			return &AppNotFoundError{options.App, ErrorTrace{}}
		}
		if len(state.migrations) == 0 {
			return &NoAppMigrationsError{options.App, ErrorTrace{}}
		}
		var node *Node
		if options.Node == "" {
			node = state.migrations[len(state.migrations)-1]
		} else {
			number, err := strconv.Atoi(options.Node[:4])
			if err != nil {
				return &NameError{options.Node, ErrorTrace{}}
			}
			if number > 0 {
				node = state.migrations[number-1]
			}
		}
		var err error
		if node == nil {
			err = state.migrations[0].Backwards(db)
		} else if node.number() < state.lastApplied {
			err = state.migrations[node.number()+1].Backwards(db)
		} else {
			err = node.Run(db)
		}
		if dbErr, ok := err.(*gomodels.DatabaseError); ok {
			dbErr.Name = dbName
			return dbErr
		} else if err != nil {
			return err
		}
	} else {
		for _, state := range history {
			if len(state.migrations) > 0 {
				node := state.migrations[len(state.migrations)-1]
				if err := node.Run(db); err != nil {
					if dbErr, ok := err.(*gomodels.DatabaseError); ok {
						dbErr.Name = dbName
						return dbErr
					}
					return err
				}
			}
		}
	}
	return nil
}
