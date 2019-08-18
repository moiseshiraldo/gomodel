package gomodels

import (
	"fmt"
	"go/build"
	"path/filepath"
)

type AppSettings struct {
	Name   string
	Path   string
	Models []*Model
}

func NewApp(name string, path string, models ...*Model) AppSettings {
	return AppSettings{name, path, models}
}

type Application struct {
	name   string
	path   string
	models map[string]*Model
}

func (app Application) Models() map[string]*Model {
	models := map[string]*Model{}
	for name, model := range app.models {
		models[name] = model
	}
	return models
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

var registry = map[string]*Application{}

func Registry() map[string]*Application {
	regCopy := map[string]*Application{}
	for name, app := range registry {
		regCopy[name] = app
	}
	return regCopy
}

func ClearRegistry() {
	registry = map[string]*Application{}
}

func Register(apps ...AppSettings) error {
	for _, settings := range apps {
		appName := settings.Name
		if _, found := registry[appName]; found || appName == "gomodels" {
			panic(fmt.Sprintf("gomodels: duplicate app: %s", settings.Name))
		}
		app := &Application{
			name:   settings.Name,
			path:   settings.Path,
			models: make(map[string]*Model),
		}
		registry[app.name] = app
		for _, model := range settings.Models {
			if err := model.Register(app); err != nil {
				panic(fmt.Sprintf(
					"gomodels: %s: %s: %s", appName, model.name, err,
				))
			}
			app.models[model.name] = model
		}
	}
	return nil
}
