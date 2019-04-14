package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var mFileRe = regexp.MustCompile(`^([0-9]{4})_\w+\.json$`)

type appState struct {
	models     map[string]*gomodels.Model
	migrations []*MigrationInfo
}

var history = map[string]*appState{}

func loadHistory() error {
	for _, app := range gomodels.Registry {
		state := &appState{
			migrations: []*MigrationInfo{},
		}
		history[app.Name()] = state
		dir := filepath.Join(app.FullPath(), MigrationsDir)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return fmt.Errorf(
				"load history failed: %v", err,
			)
		}
		for _, file := range files {
			if !mFileRe.MatchString(file.Name()) {
				return fmt.Errorf(
					"load history: invalid file name: %s", file.Name(),
				)
			}
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			m := &MigrationInfo{
				Name: name,
				Path: dir,
			}
			err := m.Load()
			if err != nil {
				return fmt.Errorf(
					"load history failed: %v", err,
				)
			}
			state.migrations = append(state.migrations, m)
			fmt.Printf("%+v\n", m)
		}
	}
	return nil
}
