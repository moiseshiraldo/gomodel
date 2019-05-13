package migrations

import (
	"database/sql"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var mFileRe = regexp.MustCompile(`^([0-9]{4})_\w+\.json$`)
var mNameRe = regexp.MustCompile(`^([0-9]{4})_\w+$`)

type AppState struct {
	Models      map[string]*gomodels.Model
	migrations  []*Node
	lastApplied int
}

func (state AppState) nextMigrationFilename(name string) string {
	if len(state.migrations) == 0 {
		return "0001_initial"
	}
	number := len(state.migrations)
	if name == "" {
		name = "auto_" + time.Now().Format("20060102_1504")
	}
	return fmt.Sprintf("%04d_%s", number+1, name)
}

var history = map[string]*AppState{}

func loadHistory() error {
	for _, app := range gomodels.Registry {
		if err := loadApp(app); err != nil {
			return err
		}
	}
	stash := map[string]map[string]bool{}
	for app := range history {
		stash[app] = map[string]bool{}
	}
	for _, state := range history {
		for _, node := range state.migrations {
			if err := node.setState(stash); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadApp(app *gomodels.Application) error {
	state := &AppState{
		Models:     map[string]*gomodels.Model{},
		migrations: []*Node{},
	}
	history[app.Name()] = state
	dir := filepath.Join(app.FullPath(), MigrationsDir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return &PathError{ErrorTrace{Err: err}}
	}
	state.migrations = make([]*Node, len(files))
	for _, file := range files {
		if !mFileRe.MatchString(file.Name()) {
			return &NameError{file.Name(), ErrorTrace{}}
		}
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		node := &Node{
			App:  app.Name(),
			Name: name,
			Path: dir,
		}
		number := node.number()
		if err := node.Load(); err != nil {
			return &LoadError{ErrorTrace{Node: node}}
		}
		if dup := state.migrations[number-1]; dup != nil {
			return &DuplicateNumberError{ErrorTrace{Node: node}}
		}
		state.migrations[number-1] = node
	}
	return nil
}

func loadApplied(db *sql.DB) error {
	rows, err := db.Query("SELECT app, number FROM gomodels_migration")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var appName string
		var number int
		err := rows.Scan(&appName, &number)
		if err != nil {
			return err
		}
		if app, ok := history[appName]; ok {
			if number <= len(app.migrations) {
				app.migrations[number-1].applied = true
			}
			if number > app.lastApplied {
				app.lastApplied = number
			}
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}
