package gomodels

import (
	"fmt"
)

type Database struct {
	ENGINE string
	NAME   string
	HOST   string
}

type Settings struct {
	Databases []Database
}

var Registry = map[string]*Application{}

func Register(settings AppSettings, appModels ...*Model) error {
	app := &Application{
		name:   settings.Name,
		path:   settings.Path,
		models: make(map[string]*Model),
	}
	Registry[settings.Name] = app
	for _, model := range appModels {
		if err := registerModel(app, model); err != nil {
			return fmt.Errorf(
				"gomodels: %s: %s: %v", settings.Name, model.name, err,
			)
		}
		Registry[settings.Name].models[model.name] = model
	}
	return nil
}

func registerModel(app *Application, model *Model) error {
	model.app = app
	for name, field := range model.fields {
		if field.IsPk() && model.pk != "" {
			return fmt.Errorf("%s: duplicate primary key", name)
		} else if field.IsPk() {
			model.pk = name
		}
	}
	if model.pk == "" {
		model.fields["id"] = AutoField{PrimaryKey: true}
		model.pk = "id"
	}
	return nil
}
