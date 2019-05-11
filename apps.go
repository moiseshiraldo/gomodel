package gomodels

import (
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
