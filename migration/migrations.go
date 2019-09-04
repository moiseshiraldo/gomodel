/*
Package migration provides the tools to detect and manage the changes made to
application models, store them in version control and apply them to the
database schema.

The Make function detects and returns all the changes made to the models in
the specified application, and optionally writes them to the project source
in JSON format.

The Run function loads the changes from the project source and apply them
to the database schema.

Alternatively, the MakeAndRun function can be used to directly create all the
models on the database schema without storing changes in the project source.
*/
package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"regexp"
)

// FileNameRegex is a regular expression to validate migration file names
// (e.g. 0001_initial.json).
var FileNameRegex = regexp.MustCompile(`^([0-9]{4})_\w+\.json$`)

// NodeNameRegex is a regular expression to validate node names
// (e.g. 0001_initial).
var NodeNameRegex = regexp.MustCompile(`^([0-9]{4})_\w+$`)

// MakeOptions holds the options for the Make function.
type MakeOptions struct {
	Empty     bool // Empty option is used to create an empty migration file.
	OmitWrite bool // OmitWrite options is used to skip writing the file.
}

// Make detects the changes between the gomodel.Model definitions and the
// migration files for the application named by appName.
//
// It returns the *AppState containing the migrations and writes them to files
// depending on MakeOptions.
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
			if err := node.Save(); err != nil {
				return state, err
			}
		}
	}
	return state, nil
}

// RunOptions holds the options for the Run function.
type RunOptions struct {
	App      string // Application name or blank for all applications.
	Node     string // Node name (e.g. 0001_inital or just 0001).
	Fake     bool   // Fake is used to apply the migration without DB changes.
	Database string // Database name or blank for default.
}

// Run applies the migrations specified by RunOptions to the database schema.
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
		return err
	}
	if options.App != "" {
		state, ok := history[options.App]
		if !ok {
			return &AppNotFoundError{options.App, ErrorTrace{}}
		}
		migrate := state.Migrate
		if options.Fake {
			migrate = state.Fake
		}
		if err := migrate(dbName, options.Node); err != nil {
			return err
		}
	} else {
		for _, state := range history {
			if len(state.migrations) > 0 {
				migrate := state.Migrate
				if options.Fake {
					migrate = state.Fake
				}
				if err := migrate(dbName, ""); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// MakeAndRun propagates all the application models to the database schema.
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
