package gomodels

import (
	"fmt"
	"go/build"
	"path/filepath"
)

type Database struct {
	ENGINE string
	NAME   string
	HOST   string
}

type Settings struct {
	Databases []Database
}

type AppSettings struct {
	Name string
	Path string
}

type Application struct {
	name   string
	path   string
	models map[string]*Model
}

func (app Application) Models() map[string]*Model {
	return app.models
}

func (app Application) Name() string {
	return app.name
}

func (app Application) Path() string {
	return app.path
}

func (app Application) FullPath() string {
	if !filepath.IsAbs(app.path) {
		return filepath.Join(build.Default.GOPATH, "src", app.path)
	}
	return app.path
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
