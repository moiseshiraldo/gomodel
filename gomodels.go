package gomodels

type Model struct {
	app    *Application
	name   string
	pk     string
	fields Fields
}

func (m Model) Name() string {
	return m.name
}

func (m Model) App() *Application {
	return m.app
}

func (m Model) Fields() Fields {
	fields := Fields{}
	for name, field := range m.fields {
		fields[name] = field
	}
	return fields
}

func New(name string, fields Fields) *Model {
	return &Model{name: name, fields: fields}
}

var Registry = map[string]*Application{}

func Register(settings AppSettings, appModels ...*Model) error {
	app := &Application{
		name:   settings.Name,
		path:   settings.Path,
		models: make(map[string]*Model),
	}
	if _, found := Registry[settings.Name]; found {
		return &DuplicateAppError{ErrorTrace{App: app}}
	}
	Registry[settings.Name] = app
	for _, model := range appModels {
		if err := registerModel(app, model); err != nil {
			return err
		}
		Registry[settings.Name].models[model.name] = model
	}
	return nil
}

func registerModel(app *Application, model *Model) error {
	if _, found := app.models[model.name]; found {
		return &DuplicateModelError{ErrorTrace{App: app, Model: model}}
	}
	model.app = app
	for name, field := range model.fields {
		if field.IsPk() && model.pk != "" {
			return &DuplicatePkError{ErrorTrace{app, model, name, nil}}
		} else if field.IsPk() {
			model.pk = name
		}
	}
	if model.pk == "" {
		model.fields["id"] = AutoField{PrimaryKey: true}
		model.pk = "id"
	}
	app.models[model.name] = model
	return nil
}
