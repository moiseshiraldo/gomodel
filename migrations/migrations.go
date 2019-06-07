package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
)

const MigrationsDir = "migrations"

type MakeOptions struct {
	Empty     bool
	OmitWrite bool
}

func Make(appName string, options MakeOptions) (*AppState, error) {
	var err error
	migrations := []*Node{}
	_, ok := gomodels.Registry[appName]
	if !ok {
		return nil, &AppNotFoundError{appName, ErrorTrace{}}
	}
	if err := loadHistory(); err != nil {
		return nil, err
	}
	defer clearHistory()
	state := history[appName]
	if options.Empty {
		node := state.nextNode()
		migrations = append(migrations, node)
		return state, nil
	}
	migrations, err = state.makeMigrations()
	if err != nil {
		return state, err
	}
	if !options.OmitWrite {
		for _, node := range migrations {
			if node.Path == "" {
				trace := ErrorTrace{Node: node, Err: fmt.Errorf("no path")}
				return state, &PathError{trace}
			}
			if err := node.Save(); err != nil {
				err = &SaveError{ErrorTrace{Node: node, Err: err}}
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
	db, ok := gomodels.Databases[dbName]
	if !ok {
		return &gomodels.DatabaseError{dbName, gomodels.ErrorTrace{}}
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

func MakeAndRun() error {
	for _, app := range gomodels.Registry {
		history[app.Name()] = &AppState{
			app:        app,
			models:     map[string]*gomodels.Model{},
			migrations: []*Node{},
		}
	}
	defer clearHistory()
	for _, state := range history {
		if _, err := state.makeMigrations(); err != nil {
			return err
		}
	}
	for dbName, db := range gomodels.Databases {
		if err := loadAppliedMigrations(db); err != nil {
			return &gomodels.DatabaseError{
				dbName, gomodels.ErrorTrace{Err: err},
			}
		}
		for _, state := range history {
			if err := state.Migrate(dbName, ""); err != nil {
				return err
			}
		}
	}
	return nil
}
