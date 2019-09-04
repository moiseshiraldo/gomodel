package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// history holds a global registry of application states.
var history = map[string]*AppState{}

// AppState holds the applicaton state for a certain node of the changes graph.
type AppState struct {
	app *gomodel.Application
	// Models holds the model definitions for this application state.
	Models      map[string]*gomodel.Model
	migrations  []*Node
	lastApplied int
}

// nextNode returns an empty Node with the next node number for this app.
func (state AppState) nextNode() *Node {
	node := &Node{
		App:          state.app.Name(),
		Path:         state.app.FullPath(),
		Dependencies: [][]string{},
		Operations:   OperationList{},
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

// MakeMigrations returns a list of nodes containing the changes between the
// gomodel.Model definitions and the migrations files for this application.
func (state *AppState) MakeMigrations() ([]*Node, error) {
	appStash := make(map[string]bool)
	return state.makeMigrations(appStash)
}

// The stash argument holds a registry of applications to keep track of
// circular dependencies.
func (state *AppState) makeMigrations(stash map[string]bool) ([]*Node, error) {
	app := state.app
	stash[app.Name()] = true
	migrations := []*Node{}
	node := state.nextNode()
	for name := range state.Models {
		// Checks for deleted models.
		if _, ok := app.Models()[name]; !ok {
			node.Operations = append(node.Operations, DeleteModel{Name: name})
		}
	}
	for _, model := range app.Models() {
		modelState, ok := state.Models[model.Name()]
		if !ok {
			// New model.
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
				// Checks for removed indexes.
				if _, ok := model.Indexes()[idxName]; !ok {
					operation := RemoveIndex{model.Name(), idxName}
					node.Operations = append(node.Operations, operation)
				}
			}
			newFields := gomodel.Fields{}
			removedFields := []string{}
			for name := range modelState.Fields() {
				// Checks for removed fields.
				if _, ok := model.Fields()[name]; !ok {
					removedFields = append(removedFields, name)
				}
			}
			if len(removedFields) > 0 {
				operation := RemoveFields{model.Name(), removedFields}
				node.Operations = append(node.Operations, operation)
			}
			for name, field := range model.Fields() {
				// Chekcs for new fields.
				if _, ok := modelState.Fields()[name]; !ok {
					newFields[name] = field
				}
			}
			if len(newFields) > 0 {
				operation := AddFields{Model: model.Name(), Fields: newFields}
				node.Operations = append(node.Operations, operation)
			}
			for idxName, fields := range model.Indexes() {
				// Checks for new indexes.
				if _, ok := modelState.Indexes()[idxName]; !ok {
					operation := AddIndex{model.Name(), idxName, fields}
					node.Operations = append(node.Operations, operation)
				}
			}
		}
	}
	if len(node.Operations) > 0 {
		nodeStash := map[string]map[string]bool{}
		for app := range history {
			nodeStash[app] = map[string]bool{}
		}
		// Applies detected changes to the application state.
		if err := node.setState(nodeStash); err != nil {
			return migrations, err
		}
		migrations = append(migrations, node)
		state.migrations = append(state.migrations, node)
	}
	delete(stash, state.app.Name())
	return migrations, nil
}

// Migrate applies the changes up to and including the node named by nodeName,
// using the db schema named by database.
//
// It will migrate all the nodes if nodeName is blank.
//
// It will run backwards if the given nodeName precedws the current applied
// node.
func (state AppState) Migrate(database string, nodeName string) error {
	return state.migrate(database, nodeName, false)
}

// Fake fakes the changes up to and including the node named by nodeName,
// for the db schema named by database.
func (state AppState) Fake(database string, nodeName string) error {
	return state.migrate(database, nodeName, true)
}

// migrate applies changes to the database schema up to and including the node
// given by name. It will skip db operations if fake is true.
func (state AppState) migrate(database string, name string, fake bool) error {
	if len(state.migrations) == 0 {
		return &NoAppMigrationsError{state.app.Name(), ErrorTrace{}}
	}
	db, ok := gomodel.Databases()[database]
	if !ok {
		return &gomodel.DatabaseError{database, gomodel.ErrorTrace{}}
	}
	var node *Node
	var backwards bool
	if name == "" {
		// Migrates to the last node.
		node = state.migrations[len(state.migrations)-1]
	} else {
		number, err := strconv.Atoi(name[:4])
		if err != nil || number > len(state.migrations) {
			return &NameError{name, ErrorTrace{}}
		}
		node = state.migrations[0]
		if number < state.lastApplied {
			backwards = true
			node = state.migrations[number]
		} else if number > 0 {
			node = state.migrations[number-1]
		}
	}
	run := node.Run
	if backwards && fake {
		run = node.FakeBackwards
	} else if backwards {
		run = node.Backwards
	} else if fake {
		run = node.Fake
	}
	return run(db)
}

// loadHistory reads the migration files and stores the application states in
// the global variable history.
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

// clearHistory clears the global application states registry.
func clearHistory() {
	history = map[string]*AppState{}
}

// loadApp loads the migration files fot the given applicaton and stores the
// application state on the global history registry.
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
		if !FileNameRegex.MatchString(name) {
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

// loadPreviousState returns the application state containing the model
// definitions up to the given node of the changes graph.
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

// loadAppliedMigrations loads the applied migrations on the given db schema.
var loadAppliedMigrations = func(db gomodel.Database) error {
	if Migration.Model.App() == nil {
		app := gomodel.Registry()["gomodel"]
		Migration.Model.Register(app)
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
		if state, ok := history[appName]; ok {
			if int(number) > len(state.migrations) {
				trace := gomodel.ErrorTrace{
					App: state.app,
					Err: fmt.Errorf("missing node for applied migration"),
				}
				return &gomodel.DatabaseError{db.Id(), trace}
			}
			state.migrations[number-1].applied = true
			if int(number) > state.lastApplied {
				state.lastApplied = int(number)
			}
		}
	}
	return nil
}
