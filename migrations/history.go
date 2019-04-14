package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type appState struct {
	models        map[string]*gomodels.Model
	lastMigration string
}

var history = map[string]*appState{}

func loadHistory() error {
	for _, app := range gomodels.Registry {
		history[app.Name()] = &appState{}
		dir := filepath.Join(app.FullPath(), MigrationsDir)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return fmt.Errorf(
				"load history failed: %v", err,
			)
		}
		for _, file := range files {
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			history[app.Name()].lastMigration = name
		}
	}
	return nil
}
