package migrations

import (
	"encoding/json"
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

const MigrationsDir = "migrations"

type MigrationInfo struct {
	App          string
	Path         string `json:"-"`
	Name         string `json:"-"`
	Dependencies []string
	Operations   OperationList
}

func (m *MigrationInfo) Save() error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	fp := filepath.Join(m.Path, m.Name+".json")
	if err := ioutil.WriteFile(fp, data, 0644); err != nil {
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	return nil
}

func (m *MigrationInfo) Load() error {
	fp := filepath.Join(m.Path, m.Name+".json")
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	if err := json.Unmarshal(data, m); err != nil {
		return fmt.Errorf("%s: %v", m.Name, err)
	}
	return nil
}

func Make(appName string) ([]*MigrationInfo, error) {
	migrations := []*MigrationInfo{}
	app, ok := gomodels.Registry[appName]
	if !ok {
		return migrations, fmt.Errorf(
			"migrations: %s: app doesn't exist", appName,
		)
	}
	if err := loadHistory(); err != nil {
		return migrations, fmt.Errorf("migrations: %v", err)
	}
	for _, model := range app.Models() {
		migrations = append(migrations, getModelChanges(model)...)
	}
	for _, m := range migrations {
		if err := m.Save(); err != nil {
			return migrations, fmt.Errorf("migrations: %s: %v", appName, err)
		}
	}
	return migrations, nil
}

func getModelChanges(model *gomodels.Model) []*MigrationInfo {
	migrations := []*MigrationInfo{}
	migration := &MigrationInfo{
		App:          model.App().Name(),
		Path:         filepath.Join(model.App().FullPath(), MigrationsDir),
		Dependencies: []string{},
	}
	operation := CreateModel{
		Model:  model.Name(),
		Fields: model.Fields(),
	}
	migration.Operations = append(migration.Operations, operation)
	migration.Name, _ = getNextMigrationName(model.App().Name())
	return append(migrations, migration)
}

func getNextMigrationName(app string) (name string, err error) {
	appHistory := history[app]
	if len(appHistory.migrations) == 0 {
		return "0001_initial", nil
	}
	lastMigration := appHistory.migrations[len(appHistory.migrations)-1]
	number, _ := strconv.Atoi(lastMigration.Name[:4])
	timestamp := time.Now().Format("20060102_1504")
	return fmt.Sprintf("%04d_auto_%s", number+1, timestamp), nil
}
