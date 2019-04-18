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

func (state AppState) nextMigrationName() string {
	if len(state.migrations) == 0 {
		return "0001_initial"
	}
	number := len(state.migrations)
	timestamp := time.Now().Format("20060102_1504")
	return fmt.Sprintf("%04d_auto_%s", number+1, timestamp)
}

var history = map[string]*AppState{}

func loadHistory() error {
	for _, app := range gomodels.Registry {
		err := loadApp(app)
		if err != nil {
			return fmt.Errorf("load history: %v", err)
		}
	}
	stash := map[string]map[string]bool{}
	for app := range history {
		stash[app] = map[string]bool{}
	}
	for _, state := range history {
		for _, node := range state.migrations {
			if err := processNode(node, stash); err != nil {
				return fmt.Errorf("load history: %v", err)
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
		return fmt.Errorf("%s: %v", app.Name(), err)
	}
	state.migrations = make([]*Node, len(files))
	for _, file := range files {
		if !mFileRe.MatchString(file.Name()) {
			return fmt.Errorf(
				"%s: invalid migration name: %s", app.Name(), file.Name(),
			)
		}
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		node := &Node{
			Name: name,
			Path: dir,
		}
		number := node.number()
		if err := node.Load(); err != nil {
			return fmt.Errorf("%s: %v", app.Name(), err)
		}
		if dup := state.migrations[number-1]; dup != nil {
			return fmt.Errorf("%s: duplicate number: %s", app.Name(), name)
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
	for _, dep := range node.Dependencies {
		app, name := dep[0], dep[1]
		if !mNameRe.MatchString(name) {
			return fmt.Errorf("invalid dependency: %s", name)
		}
		number, _ := strconv.Atoi(name[:4])
		if number > len(history[node.App].migrations) {
			return fmt.Errorf("invalid dependency: %s", name)
		}
		depNode := history[node.App].migrations[number-1]
		if depNode == nil {
			return fmt.Errorf("invalid dependency: %s", name)
		}
		if _, found := stash[node.App][name]; found {
			return fmt.Errorf("circular dependency: %s: %s", app, depNode.Name)
		}
		if !depNode.processed {
			if err := processNode(depNode, stash); err != nil {
				return err
			}
		}
	}
	for _, op := range node.Operations {
		if err := op.SetState(history[node.App]); err != nil {
			return fmt.Errorf("%s: set state: %v", node.Name, err)
		}
	}
	node.processed = true
	delete(stash[node.App], node.Name)
	return nil
}
