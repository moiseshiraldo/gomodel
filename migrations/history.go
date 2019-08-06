package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var mFileRe = regexp.MustCompile(`^([0-9]{4})_\w+\.json$`)
var mNameRe = regexp.MustCompile(`^([0-9]{4})_\w+$`)

var history = map[string]*AppState{}

type AppState struct {
	app         *gomodels.Application
	models      map[string]*gomodels.Model
	migrations  []*Node
	lastApplied int
}

func (state AppState) nextNode() *Node {
	node := &Node{
		App:          state.app.Name(),
		Dependencies: [][]string{},
		Operations:   OperationList{},
	}
	if state.app.Path() != "" {
		node.Path = state.app.FullPath()
	}
	node.number = len(state.migrations) + 1
	if node.number == 1 {
		node.Name = "initial"
	} else {
		timestamp := time.Now().Format("20060102_1504")
		node.Name = fmt.Sprintf("auto_%s", timestamp)
	}
	if len(state.migrations) > 0 {
		lastNode := state.migrations[len(state.migrations)-1]
		node.Dependencies = append(
			node.Dependencies, []string{state.app.Name(), lastNode.fullname()},
		)
	}
	return node
}

func (state *AppState) makeMigrations() ([]*Node, error) {
	migrations := []*Node{}
	node := state.nextNode()
	for name := range state.models {
		if _, ok := state.app.Models()[name]; !ok {
			node.Operations = append(node.Operations, &DeleteModel{Name: name})
		}
	}
	for _, model := range state.app.Models() {
		node.Operations = append(node.Operations, getModelChanges(model)...)
	}
	if len(node.Operations) > 0 {
		stash := map[string]map[string]bool{}
		for app := range history {
			stash[app] = map[string]bool{}
		}
		if err := node.setState(stash); err != nil {
			return migrations, err
		}
		migrations = append(migrations, node)
		state.migrations = append(state.migrations, node)
	}
	return migrations, nil
}

func (state AppState) Migrate(database string, nodeName string) error {
	if len(state.migrations) == 0 {
		return &NoAppMigrationsError{state.app.Name(), ErrorTrace{}}
	}
	db, ok := gomodels.Databases()[database]
	if !ok {
		return &gomodels.DatabaseError{database, gomodels.ErrorTrace{}}
	}
	var node *Node
	if nodeName == "" {
		node = state.migrations[len(state.migrations)-1]
	} else {
		number, err := strconv.Atoi(nodeName[:4])
		if err != nil || number > len(state.migrations) {
			return &NameError{nodeName, ErrorTrace{}}
		}
		if number > 0 {
			node = state.migrations[number-1]
		}
	}
	var err error
	if node == nil {
		err = state.migrations[0].Backwards(db)
	} else if node.number < state.lastApplied {
		err = state.migrations[node.number].Backwards(db)
	} else {
		err = node.Run(db)
	}
	if dbErr, ok := err.(*gomodels.DatabaseError); ok {
		dbErr.Name = database
		return dbErr
	} else if err != nil {
		return err
	}
	return nil
}

func loadHistory() error {
	for _, app := range gomodels.Registry() {
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

func clearHistory() {
	history = map[string]*AppState{}
}

func loadApp(app *gomodels.Application) error {
	state := &AppState{
		app:        app,
		models:     map[string]*gomodels.Model{},
		migrations: []*Node{},
	}
	history[app.Name()] = state
	if app.Path() == "" {
		return nil
	}
	dir := app.FullPath()
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return &PathError{ErrorTrace{Err: err}}
	}
	state.migrations = make([]*Node, len(files))
	for _, file := range files {
		if !mFileRe.MatchString(file.Name()) {
			return &NameError{file.Name(), ErrorTrace{}}
		}
		filename := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		number, _ := strconv.Atoi(filename[:4])
		node := &Node{
			App:    app.Name(),
			Name:   filename[5:],
			number: number,
			Path:   dir,
		}
		if err := node.Load(); err != nil {
			return &LoadError{ErrorTrace{Node: node, Err: err}}
		}
		if dup := state.migrations[number-1]; dup != nil {
			return &DuplicateNumberError{ErrorTrace{Node: node}}
		}
		state.migrations[number-1] = node
	}
	return nil
}

func loadPreviousState(node Node) map[string]*AppState {
	prevState := map[string]*AppState{}
	registry := gomodels.Registry()
	for name := range history {
		prevState[name] = &AppState{
			app:    registry[node.App],
			models: map[string]*gomodels.Model{},
		}
	}
	if node.number > 1 {
		prevNode := history[node.App].migrations[node.number-2]
		prevNode.setPreviousState(prevState)
	}
	return prevState
}

func loadAppliedMigrations(db gomodels.Database) error {
	if err := db.PrepareMigrations(); err != nil {
		return err
	}
	rows, err := db.GetMigrations()
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
