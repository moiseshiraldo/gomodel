package gomodels

import (
	"fmt"
	"strings"
)

type Dispatcher struct {
	*Model
	Objects *Manager
}

type Model struct {
	app    *Application
	name   string
	pk     string
	fields Fields
	meta   Options
}

type Indexes map[string][]string

type Options struct {
	Constructor Constructor
	Indexes     Indexes
}

func (m Model) Name() string {
	return m.name
}

func (m Model) App() *Application {
	return m.app
}

func (m Model) Table() string {
	return fmt.Sprintf(
		"%s_%s", strings.ToLower(m.app.name), strings.ToLower(m.name),
	)
}

func (m Model) Fields() Fields {
	fields := Fields{}
	for name, field := range m.fields {
		fields[name] = field
	}
	return fields
}

func (m Model) Indexes() Indexes {
	indexes := Indexes{}
	for name, fields := range m.meta.Indexes {
		fieldsCopy := make([]string, len(fields))
		copy(fieldsCopy, fields)
		indexes[name] = fieldsCopy
	}
	return indexes
}

func New(name string, fields Fields, options Options) *Dispatcher {
	if options.Indexes == nil {
		options.Indexes = Indexes{}
	}
	model := &Model{name: name, fields: fields, meta: options}
	return &Dispatcher{model, &Manager{model}}
}

var Registry = map[string]*Application{}

func Register(apps ...AppSettings) error {
	for _, settings := range apps {
		appName := settings.Name
		if _, found := Registry[appName]; found || appName == "gomodels" {
			panic(fmt.Sprintf("gomodels: duplicate app: %s", settings.Name))
		}
		app := &Application{
			name:   settings.Name,
			path:   settings.Path,
			models: make(map[string]*Model),
		}
		Registry[app.name] = app
		for _, model := range settings.Models {
			registerModel(app, model)
			Registry[app.name].models[model.name] = model
		}
	}
	return nil
}

func registerModel(app *Application, model *Model) {
	if _, found := app.models[model.name]; found {
		panic(fmt.Sprintf(
			"gomodels: %s: duplicate model: %s", app.name, model.name,
		))
	}
	model.app = app
	for name, field := range model.fields {
		if field.IsPk() && model.pk != "" {
			msg := fmt.Sprintf(
				"gomodels: %s: %s: %s: duplicate pk",
				app.name, model.name, name,
			)
			panic(msg)
		} else if field.IsPk() {
			model.pk = name
		}
		if field.HasIndex() {
			idxName := fmt.Sprintf(
				"%s_%s_%s_idx", app.name, model.name, field.DBColumn(name),
			)
			if _, found := model.meta.Indexes[idxName]; found {
				msg := fmt.Sprintf(
					"gomodels: %s: %s: duplicate index: %s",
					app.name, model.name, idxName,
				)
				panic(msg)
			}
			model.meta.Indexes[idxName] = []string{field.DBColumn(name)}
		}
	}
	if model.pk == "" {
		model.fields["id"] = AutoField{PrimaryKey: true}
		model.pk = "id"
	}
	if model.meta.Constructor != nil {
		if getConstructorType(model.meta.Constructor) == "" {
			panic(fmt.Sprintf(
				"gomodels: %s: %s: invalid constructor", app.name, model.name,
			))
		}
	}
	app.models[model.name] = model
}
