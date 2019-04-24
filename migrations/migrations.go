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

type RunOptions struct {
	App      string
	Name     string
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
	if err := loadHistory(); err != nil {
		return err
	}
	if options.App != "" {
		state, ok := history[options.App]
		if !ok {
			return &AppNotFoundError{Name: options.App}
		}
		node := &Node{}
		if options.Name != "" {
			number, err := strconv.Atoi(options.Name[:4])
			if err != nil {
				app := gomodels.Registry[options.App]
				return &NameError{
					options.Name, gomodels.ErrorTrace{App: app, Err: err},
				}
			}
			node = state.migrations[number-1]
		} else if len(state.migrations) > 0 {
			node = state.migrations[len(state.migrations)-1]
		}
		if err := node.Run(db); err != nil {
			if dbErr, ok := err.(*gomodels.DatabaseError); ok {
				dbErr.Name = dbName
				return dbErr
			}
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
