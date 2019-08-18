package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

type Dispatcher struct {
	*Model
	Objects *Manager
}

func (d Dispatcher) New(values Values) (*Instance, error) {
	model := d.Model
	instance := &Instance{model, model.meta.Container}
	for name, field := range model.fields {
		if val, ok := values[name]; ok {
			if err := instance.Set(name, val); err != nil {
				return nil, err
			}
		} else if val, hasDefault := field.DefaultVal(); hasDefault {
			if err := instance.Set(name, val); err != nil {
				return nil, err
			}
		}
	}
	return instance, nil
}

type Indexes map[string][]string

type Options struct {
	Table     string
	Container Container
	Indexes   Indexes
}

type Model struct {
	app    *Application
	name   string
	pk     string
	fields Fields
	meta   Options
}

func (m Model) Name() string {
	return m.name
}

func (m Model) App() *Application {
	return m.app
}

func (m Model) Table() string {
	table := m.meta.Table
	if table == "" {
		table = fmt.Sprintf(
			"%s_%s", strings.ToLower(m.app.name), strings.ToLower(m.name),
		)
	}
	return table
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

func (m Model) Container() Container {
	if b, ok := m.meta.Container.(Builder); ok {
		return b.New()
	} else {
		ct := reflect.TypeOf(m.meta.Container)
		if ct.Kind() == reflect.Ptr {
			ct = ct.Elem()
		}
		return reflect.New(ct).Interface()
	}
}

func (m *Model) Register(app *Application) error {
	if _, found := app.models[m.name]; found {
		return fmt.Errorf("duplicate model")
	}
	m.app = app
	if err := m.SetupPrimaryKey(); err != nil {
		return err
	}
	if err := m.SetupIndexes(); err != nil {
		return err
	}
	if m.meta.Container != nil {
		if !isValidContainer(m.meta.Container) {
			return fmt.Errorf("invalid container")
		}
	} else {
		m.meta.Container = Values{}
	}
	app.models[m.name] = m
	return nil
}

func (m *Model) SetupPrimaryKey() error {
	if m.pk != "" {
		return nil
	}
	for name, field := range m.fields {
		if field.IsPk() && m.pk != "" {
			return fmt.Errorf("duplicate pk: %s", name)
		} else if field.IsPk() {
			m.pk = name
		}
	}
	if m.pk == "" {
		m.fields["id"] = AutoField{PrimaryKey: true}
		m.pk = "id"
	}
	return nil
}

func (m *Model) SetupIndexes() error {
	for name, field := range m.fields {
		if field.HasIndex() {
			idxName := fmt.Sprintf(
				"%s_%s_%s_auto_idx",
				strings.ToLower(m.app.name),
				strings.ToLower(m.name),
				strings.ToLower(name),
			)
			if _, found := m.meta.Indexes[idxName]; found {
				return fmt.Errorf("duplicate index: %s", idxName)
			}
			m.meta.Indexes[idxName] = []string{name}
		}
	}
	return nil
}

func (m *Model) AddField(name string, field Field) error {
	if _, found := m.fields[name]; found {
		return fmt.Errorf("duplicate field: %s", name)
	}
	if field.IsPk() && m.pk != "" {
		return fmt.Errorf("duplicate pk: %s", name)
	}
	m.fields[name] = field
	if field.IsPk() {
		m.pk = name
	}
	return nil
}

func (m *Model) RemoveField(name string) error {
	if m.pk == name {
		return fmt.Errorf("pk field cannot be removed")
	}
	if _, ok := m.fields[name]; !ok {
		return fmt.Errorf("field not found: %s", name)
	}
	for _, fields := range m.meta.Indexes {
		for _, indexedField := range fields {
			if name == indexedField {
				return fmt.Errorf("cannot remove indexed field: %s", name)
			}
		}
	}
	delete(m.fields, name)
	return nil
}

func (m *Model) AddIndex(name string, fields []string) error {
	if _, found := m.meta.Indexes[name]; found {
		return fmt.Errorf("duplicate index: %s", name)
	}
	for _, indexedField := range fields {
		if _, ok := m.fields[indexedField]; !ok {
			return fmt.Errorf("unknown indexed field: %s", indexedField)
		}
	}
	m.meta.Indexes[name] = fields
	return nil
}

func (m *Model) RemoveIndex(name string) error {
	if _, ok := m.meta.Indexes[name]; !ok {
		return fmt.Errorf("index not found: %s", name)
	}
	delete(m.meta.Indexes, name)
	return nil
}

func New(name string, fields Fields, options Options) *Dispatcher {
	if options.Indexes == nil {
		options.Indexes = Indexes{}
	}
	model := &Model{name: name, fields: fields, meta: options}
	return &Dispatcher{model, &Manager{model}}
}
