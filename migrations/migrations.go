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
	Operations   []Operation
}

func (m *MigrationInfo) Write() error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf(
			"%s: write failed: %v", m.App, err,
		)
	}
	fp := filepath.Join(m.Path, m.Name+".json")
	err = ioutil.WriteFile(fp, data, 0644)
	if err != nil {
		return fmt.Errorf(
			"%s: write failed: %v", m.App, err,
		)
	}
	return nil
}

func Make(appName string) ([]*MigrationInfo, error) {
	migrations := []*MigrationInfo{}
	app, ok := gomodels.Registry[appName]
	if !ok {
		return migrations, fmt.Errorf(
			"gomodels: makemigrations: %s: app doesn't exist", appName,
		)
	}
	err := loadHistory()
	if err != nil {
		return migrations, fmt.Errorf(
			"gomodels: makemigrations: %v", err,
		)
	}
	for _, model := range app.Models() {
		migrations = append(migrations, getModelChanges(model)...)
	}
	for _, m := range migrations {
		err := m.Write()
		if err != nil {
			return migrations, fmt.Errorf(
				"gomodels: makemigrations: %v", err,
			)
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
	operation := Operation{
		CreateModel: CreateModel{
			Name:   model.Name(),
			Fields: make(map[string]FieldDesc),
		},
	}
	for fieldName, field := range model.Fields() {
		operation.CreateModel.Fields[fieldName] = FieldDesc{
			Options: field,
		}
	}
	migration.Operations = append(migration.Operations, operation)
	migration.Name, _ = getNextMigrationName(model.App().Name())
	return append(migrations, migration)
}

func getNextMigrationName(app string) (name string, err error) {
	appHistory := history[app]
	if appHistory.lastMigration == "" {
		return "0001_initial", nil
	}
	number, _ := strconv.Atoi(appHistory.lastMigration[:4])
	timestamp := time.Now().Format("20060102_1504")
	return fmt.Sprintf("%04d_auto_%s", number+1, timestamp), nil
}
