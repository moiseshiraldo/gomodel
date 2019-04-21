package gomodels

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
