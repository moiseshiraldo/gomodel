package gomodels

type Field interface {
	IsPk() bool
	FromJson(raw []byte) (Field, error)
}

type Fields map[string]Field

type Model struct {
	app    *application
	name   string
	pk     string
	fields Fields
}

func (m Model) Name() string {
	return m.name
}

func (m Model) App() *application {
	return m.app
}

func (m Model) Fields() Fields {
	return m.fields
}

func New(name string, fields Fields) *Model {
	return &Model{name: name, fields: fields}
}

func AvailableFields() map[string]Field {
	return map[string]Field{
		"IntegerField": IntegerField{},
		"AutoField":    AutoField{},
		"BooleanField": BooleanField{},
		"CharField":    CharField{},
	}
}