package gomodel

import (
	"fmt"
	"go/build"
	"path/filepath"
)

// AppSettings holds the options te register a new application.
type AppSettings struct {
	// Name is the application name, which must be unique.
	Name string
	// Path must be an existing directory path or blank if migrations won't be
	// managed by the gomodel/migration package. If relative, the full path will
	// be constructed from $GOPATH/src.
	Path string
	// Models is the list of models that will be registered to the application.
	Models []*Model
}

// NewApp returns an application settings struct constructed from the given
// arguments, ready to registered using the Register function.
//
// The name is the applicatoin name, which must be unique.
//
// The path must be an existing directory path or blank if migrations won't be
// managed by the gomodel/migration package. If relative, the full path will
// be constructed from $GOPATH/src.
//
// The last argument is the list of models that will be registered to the app.
func NewApp(name string, path string, models ...*Model) AppSettings {
	return AppSettings{name, path, models}
}

// Application holds a registered application and the validated models ready
// to interact with the database.
type Application struct {
	name   string
	path   string
	models map[string]*Model
}

// Models returns a map of models registered to the application.
func (app Application) Models() map[string]*Model {
	models := map[string]*Model{}
	for name, model := range app.models {
		models[name] = model
	}
	return models
}

// Name returns the application name.
func (app Application) Name() string {
	return app.name
}

// Path returns the migrations path for the application.
func (app Application) Path() string {
	return app.path
}

// FullPath returns the full path to the migrations directory.
func (app Application) FullPath() string {
	if app.path == "" || filepath.IsAbs(app.path) {
		return app.path
	}
	return filepath.Join(build.Default.GOPATH, "src", app.path)
}

// registry hols a global map of registered applications.
var registry = map[string]*Application{}

// Registry returns a map containing all the registered applications.
func Registry() map[string]*Application {
	regCopy := map[string]*Application{}
	for name, app := range registry {
		regCopy[name] = app
	}
	return regCopy
}

// ClearRegistry removes all the registered applications.
func ClearRegistry() {
	registry = map[string]*Application{}
}

// Register validates the given application settings and their models and
// adds them to the registry. The function will panic on any validation error.
func Register(apps ...AppSettings) {
	for _, settings := range apps {
		appName := settings.Name
		if _, found := registry[appName]; found || appName == "gomodel" {
			panic(fmt.Sprintf("gomodel: duplicate app: %s", settings.Name))
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
					"gomodel: %s: %s: %s", appName, model.name, err,
				))
			}
			app.models[model.name] = model
		}
	}
}
