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

type AppState struct {
	Models     map[string]*gomodels.Model
	migrations []*Node
}

func (state AppState) nextMigrationFilename(name string) string {
	if len(state.migrations) == 0 {
		return "0001_initial"
	}
	number := len(state.migrations)
	if name != "" {
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
			if err := processNode(node, stash); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadApp(app *gomodels.Application) error {
	errorTrace := gomodels.ErrorTrace{App: app}
	state := &AppState{
		Models:     map[string]*gomodels.Model{},
		migrations: []*Node{},
	}
	history[app.Name()] = state
	dir := filepath.Join(app.FullPath(), MigrationsDir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		errorTrace.Err = err
		return &PathError{errorTrace}
	}
	state.migrations = make([]*Node, len(files))
	for _, file := range files {
		if !mFileRe.MatchString(file.Name()) {
			return &NameError{file.Name(), errorTrace}
		}
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		node := &Node{
			Name: name,
			Path: dir,
		}
		number := node.number()
		if err := node.Load(); err != nil {
			return &LoadError{file.Name(), errorTrace}
		}
		if dup := state.migrations[number-1]; dup != nil {
			return &DuplicateNumberError{file.Name(), errorTrace}
		}
		state.migrations[number-1] = node
	}
	return nil
}

func processNode(node *Node, stash map[string]map[string]bool) error {
	if node.processed {
		return nil
	}
	stash[node.App][node.Name] = true
	errorTrace := gomodels.ErrorTrace{App: gomodels.Registry[node.App]}
	for _, dep := range node.Dependencies {
		app, name := dep[0], dep[1]
		if !mNameRe.MatchString(name) {
			return &InvalidDependencyError{node.Name, name, errorTrace}
		}
		if _, ok := history[app]; !ok {
			return &InvalidDependencyError{node.Name, name, errorTrace}
		}
		number, _ := strconv.Atoi(name[:4])
		if number > len(history[app].migrations) {
			return &InvalidDependencyError{node.Name, name, errorTrace}
		}
		depNode := history[app].migrations[number-1]
		if depNode == nil {
			return &InvalidDependencyError{node.Name, name, errorTrace}
		}
		if _, found := stash[app][name]; found {
			return &CircularDependencyError{node.Name, name, errorTrace}
		}
		if !depNode.processed {
			if err := processNode(depNode, stash); err != nil {
				return err
			}
		}
	}
	for _, op := range node.Operations {
		if err := op.SetState(history[node.App]); err != nil {
			return &OperationStateError{node.Name, &op, errorTrace}
		}
	}
	node.processed = true
	delete(stash[node.App], node.Name)
	return nil
}
