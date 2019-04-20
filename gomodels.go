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
