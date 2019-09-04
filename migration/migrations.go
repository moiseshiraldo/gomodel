package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
)

type MakeOptions struct {
	Empty     bool
	OmitWrite bool
}

func Make(appName string, options MakeOptions) (*AppState, error) {
	_, ok := gomodel.Registry()[appName]
	if !ok {
		return nil, &AppNotFoundError{appName, ErrorTrace{}}
	}
	if err := loadHistory(); err != nil {
		return nil, err
	}
	defer clearHistory()
	state := history[appName]
	if options.Empty {
		state.migrations = []*Node{state.nextNode()}
		return state, nil
	}
	migrations, err := state.MakeMigrations()
	if err != nil {
		return state, err
	}
	if !options.OmitWrite {
		for _, node := range migrations {
			if node.Path == "" {
				trace := ErrorTrace{Err: fmt.Errorf("no path")}
				return state, &PathError{node.App, trace}
			}
			if err := node.Save(); err != nil {
				return state, err
			}
		}
	}
	return state, nil
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
	db, ok := gomodel.Databases()[dbName]
	if !ok {
		err := fmt.Errorf("database not found")
		return &gomodel.DatabaseError{dbName, gomodel.ErrorTrace{Err: err}}
	}
	if err := loadHistory(); err != nil {
		return err
	}
	defer clearHistory()
	if err := loadAppliedMigrations(db); err != nil {
		return &gomodel.DatabaseError{dbName, gomodel.ErrorTrace{Err: err}}
	}
	if options.App != "" {
		state, ok := history[options.App]
		if !ok {
			return &AppNotFoundError{options.App, ErrorTrace{}}
		}
		if err := state.Migrate(dbName, options.Node); err != nil {
			return err
		}
	} else {
		for _, state := range history {
			if len(state.migrations) > 0 {
				if err := state.Migrate(dbName, ""); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func MakeAndRun(database string) error {
	for _, app := range gomodel.Registry() {
		if app.Name() == "gomodel" {
			continue
		}
		history[app.Name()] = &AppState{
			app:        app,
			Models:     map[string]*gomodel.Model{},
			migrations: []*Node{},
		}
	}
	defer clearHistory()
	for _, state := range history {
		if _, err := state.MakeMigrations(); err != nil {
			return err
		}
	}
	db, ok := gomodel.Databases()[database]
	if !ok {
		err := fmt.Errorf("database not found")
		return &gomodel.DatabaseError{database, gomodel.ErrorTrace{Err: err}}
	}
	if err := loadAppliedMigrations(db); err != nil {
		return &gomodel.DatabaseError{
			database, gomodel.ErrorTrace{Err: err},
		}
	}
	for _, state := range history {
		if err := state.Migrate(database, ""); err != nil {
			return err
		}
	}
	return nil
}
