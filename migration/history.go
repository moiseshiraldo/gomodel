package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
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
	app         *gomodel.Application
	Models      map[string]*gomodel.Model
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

func (state *AppState) MakeMigrations() ([]*Node, error) {
	appStash := make(map[string]bool)
	return state.makeMigrations(appStash)
}

func (state *AppState) makeMigrations(stash map[string]bool) ([]*Node, error) {
	app := state.app
	stash[app.Name()] = true
	migrations := []*Node{}
	node := state.nextNode()
	for name := range state.Models {
		if _, ok := app.Models()[name]; !ok {
			node.Operations = append(node.Operations, DeleteModel{Name: name})
		}
	}
	for _, model := range app.Models() {
		modelState, ok := state.Models[model.Name()]
		if !ok {
			operation := CreateModel{
				Name:   model.Name(),
				Fields: model.Fields(),
			}
			defaultTable := fmt.Sprintf("%s__%s", app.Name(), model.Name())
			if model.Table() != defaultTable {
				operation.Table = model.Table()
			}
			node.Operations = append(node.Operations, operation)
			for idxName, fields := range model.Indexes() {
				operation := AddIndex{model.Name(), idxName, fields}
				node.Operations = append(node.Operations, operation)
			}
		} else {
			for idxName := range modelState.Indexes() {
				if _, ok := model.Indexes()[idxName]; !ok {
					operation := RemoveIndex{model.Name(), idxName}
					node.Operations = append(node.Operations, operation)
				}
			}
			newFields := gomodel.Fields{}
			removedFields := []string{}
			for name := range modelState.Fields() {
				if _, ok := model.Fields()[name]; !ok {
					removedFields = append(removedFields, name)
				}
			}
			if len(removedFields) > 0 {
				operation := RemoveFields{model.Name(), removedFields}
				node.Operations = append(node.Operations, operation)
			}
			for name, field := range model.Fields() {
				if _, ok := modelState.Fields()[name]; !ok {
					newFields[name] = field
				}
			}
			if len(newFields) > 0 {
				operation := AddFields{Model: model.Name(), Fields: newFields}
				node.Operations = append(node.Operations, operation)
			}
			for idxName, fields := range model.Indexes() {
				if _, ok := modelState.Indexes()[idxName]; !ok {
					operation := AddIndex{model.Name(), idxName, fields}
					node.Operations = append(node.Operations, operation)
				}
			}
		}
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
	delete(stash, state.app.Name())
	return migrations, nil
}

func (state AppState) Migrate(database string, nodeName string) error {
	if len(state.migrations) == 0 {
		return &NoAppMigrationsError{state.app.Name(), ErrorTrace{}}
	}
	db, ok := gomodel.Databases()[database]
	if !ok {
		return &gomodel.DatabaseError{database, gomodel.ErrorTrace{}}
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
	if node == nil {
		return state.migrations[0].Backwards(db)
	} else if node.number < state.lastApplied {
		return state.migrations[node.number].Backwards(db)
	} else {
		return node.Run(db)
	}
}

var loadHistory = func() error {
	for _, app := range gomodel.Registry() {
		if app.Name() == "gomodel" {
			continue
		}
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

func loadApp(app *gomodel.Application) error {
	state := &AppState{
		app:        app,
		Models:     map[string]*gomodel.Model{},
		migrations: []*Node{},
	}
	history[app.Name()] = state
	if app.Path() == "" {
		return nil
	}
	files, err := readAppNodes(app.FullPath())
	if err != nil {
		return &PathError{app.Name(), ErrorTrace{Err: err}}
	}
	for _, name := range files {
		if !mFileRe.MatchString(name) {
			return &NameError{name, ErrorTrace{}}
		}
	}
	state.migrations = make([]*Node, len(files))
	for _, name := range files {
		filename := strings.TrimSuffix(name, filepath.Ext(name))
		number, _ := strconv.Atoi(filename[:4])
		node := &Node{
			App:    app.Name(),
			Name:   filename[5:],
			number: number,
			Path:   app.FullPath(),
		}
		if dup := state.migrations[number-1]; dup != nil {
			err := fmt.Errorf("duplicate number")
			return &DuplicateNumberError{ErrorTrace{Node: node, Err: err}}
		}
		if err := node.Load(); err != nil {
			return err
		}
		state.migrations[number-1] = node
	}
	return nil
}

func loadPreviousState(node Node) map[string]*AppState {
	prevState := map[string]*AppState{}
	registry := gomodel.Registry()
	for name := range history {
		if name == "gomodel" {
			continue
		}
		prevState[name] = &AppState{
			app:    registry[node.App],
			Models: map[string]*gomodel.Model{},
		}
	}
	if node.number > 1 {
		prevNode := history[node.App].migrations[node.number-2]
		prevNode.setPreviousState(prevState)
	}
	return prevState
}

var loadAppliedMigrations = func(db gomodel.Database) error {
	if Migration.Model.App() == nil {
		app := gomodel.Registry()["gomodel"]
		if err := Migration.Model.Register(app); err != nil {
			return err
		}
	}
	if err := db.CreateTable(Migration.Model, false); err != nil {
		return err
	}
	migrations, err := Migration.Objects.All().Load()
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		appName := migration.Get("name").(string)
		number := migration.Get("number").(int32)
		if app, ok := history[appName]; ok {
			if int(number) > len(app.migrations) {
				return fmt.Errorf("missing node for applied migration")
			}
			app.migrations[number-1].applied = true
			if int(number) > app.lastApplied {
				app.lastApplied = int(number)
			}
		}
	}
	return nil
}
