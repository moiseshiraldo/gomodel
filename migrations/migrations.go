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

func Make(appName string, options MakeOptions) ([]*Node, error) {
	migrations := []*Node{}
	_, ok := gomodels.Registry[appName]
	if !ok {
		return migrations, &AppNotFoundError{appName, ErrorTrace{}}
	}
	if err := loadHistory(); err != nil {
		return migrations, err
	}
	defer clearHistory()
	state := history[appName]
	if options.Empty {
		node := state.nextNode()
		migrations = append(migrations, node)
		return migrations, nil
	}
	migrations = state.changes()
	if options.OmitWrite {
		return migrations, nil
	}
	for _, node := range migrations {
		if node.Path == "" {
			trace := ErrorTrace{Node: node, Err: fmt.Errorf("no path")}
			return migrations, &PathError{trace}
		}
		if err := node.Save(); err != nil {
			err = &SaveError{ErrorTrace{Node: node, Err: err}}
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
