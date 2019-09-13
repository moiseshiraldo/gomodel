/*
Package gomodel is an ORM that provides the resources to define data models
and a database-abstraction API that lets you create, retrieve, update and
delete objects.
*/
package gomodel

import (
	"fmt"
	"strings"
)

// A Dispatcher embedds a Model definition and holds the default Objects Manager
// that gives access to the methods to interact with the database.
type Dispatcher struct {
	*Model
	Objects Manager
}

// New returns an Instance of the embedded model, populating the fields with the
// values argument. If no value is provided for a field, it will try to get a
// default value from the model definition.
func (d Dispatcher) New(values Container) (*Instance, error) {
	model := d.Model
	instance := &Instance{model, model.meta.Container}
	for name, field := range model.fields {
		var value Value
		if val, ok := getContainerField(values, name); ok {
			value = val
		} else if val, hasDefault := field.DefaultVal(); hasDefault {
			value = val
		}
		if err := instance.Set(name, value); err != nil {
			return nil, err
		}
	}
	return instance, nil
}

// The Indexes type describes the indexes of a model, where the key is the index
// name and the value the list of indexes fields.
type Indexes map[string][]string

// Options holds extra model options.
type Options struct {
	// Table is the name of the database Table for the model. If blank, the
	// table will be {app_name}_{model_name} all lowercase.
	Table string
	// Container is a value of the type that will be used to hold the model
	// field when a new instance is created. If nil, the Values type will be
	// the default Container.
	Container Container
	// Indexes is used to declare composite indexes. Indexes with one column
	// should be defined at field level.
	Indexes Indexes
}

// A Model represents a single basic data structure of an application and how
// to map that data to the database schema.
type Model struct {
	app    *Application
	name   string
	pk     string
	fields Fields
	meta   Options
}

// Name returns the model name.
func (m Model) Name() string {
	return m.name
}

// App returns the Application containing the model.
func (m Model) App() *Application {
	return m.app
}

// Table returns the name of the database table.
func (m Model) Table() string {
	table := m.meta.Table
	if table == "" {
		table = fmt.Sprintf(
			"%s_%s", strings.ToLower(m.app.name), strings.ToLower(m.name),
		)
	}
	return table
}

// Fields returns the model Fields map.
func (m Model) Fields() Fields {
	fields := Fields{}
	for name, field := range m.fields {
		fields[name] = field
	}
	return fields
}

// Indexes returns the model Indexes map.
func (m Model) Indexes() Indexes {
	indexes := Indexes{}
	for name, fields := range m.meta.Indexes {
		fieldsCopy := make([]string, len(fields))
		copy(fieldsCopy, fields)
		indexes[name] = fieldsCopy
	}
	return indexes
}

// Container returns a new zero value of the model Container.
func (m Model) Container() Container {
	return newContainer(m.meta.Container)
}

// Register validates the model definition, calls the SetupPrimaryKey and
// SetupIndexes methods, and adds the model to the given app.
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

// SetupPrimaryKey searches the model fields for a primary key. If not found,
// it will add an auto incremented IntegerField called id.
//
// This method ia automatically called when a model is registered and should
// only be used to modify a model state during migration operations.
func (m *Model) SetupPrimaryKey() error {
	if m.pk != "" {
		return nil
	}
	for name, field := range m.fields {
		if field.IsPK() && m.pk != "" {
			return fmt.Errorf("duplicate pk: %s", name)
		} else if field.IsPK() {
			m.pk = name
		}
	}
	if m.pk == "" {
		m.fields["id"] = IntegerField{PrimaryKey: true, Auto: true}
		m.pk = "id"
	}
	return nil
}

// SetupIndexes validates the model Indexes definition and adds individually
// indexes fields.
//
// This method ia automatically called when a model is registered and should
// only be used to modify a model state during migration operations.
func (m *Model) SetupIndexes() error {
	for name, fields := range m.meta.Indexes {
		if len(fields) == 0 {
			return fmt.Errorf("index with no fields: %s", name)
		}
		for _, indexedField := range fields {
			if _, ok := m.fields[indexedField]; !ok {
				return fmt.Errorf("unknown indexed field: %s", indexedField)
			}
		}
	}
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

// AddField adds a new Field to the model definition. It returns an error if the
// field name already exists or if a duplicate primary key is added.
//
// This method should only be used to modify a model state during migration
// operations or to construct models programatically. Changing the model
// definition after it has been registered could cause unexpected errors.
func (m *Model) AddField(name string, field Field) error {
	if _, found := m.fields[name]; found {
		return fmt.Errorf("duplicate field: %s", name)
	}
	if field.IsPK() && m.pk != "" {
		return fmt.Errorf("duplicate pk: %s", name)
	}
	m.fields[name] = field
	return nil
}

// RemoveField removes the named field from the model definition. It returns an
// error if the fields is the primary key, the field is indexed or it doesn't
// exist.
//
// This method should only be used to modify a model state during migration
// operations or to construct models programatically. Changing the model
// definition after it has been registered could cause unexpected errors.
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

// AddIndex adds a new index to the model definition. It returns an error if the
// name is duplicate or any of the indexed fields doesn't exist.
//
// This method should only be used to modify a model state during migration
// operations or to construct models programatically. Changing the model
// definition after it has been registered could cause unexpected errors.
func (m *Model) AddIndex(name string, fields ...string) error {
	if _, found := m.meta.Indexes[name]; found {
		return fmt.Errorf("duplicate index: %s", name)
	}
	if len(fields) == 0 {
		return fmt.Errorf("cannot add index with no fields: %s", name)
	}
	for _, indexedField := range fields {
		if _, ok := m.fields[indexedField]; !ok {
			return fmt.Errorf("unknown indexed field: %s", indexedField)
		}
	}
	m.meta.Indexes[name] = fields
	return nil
}

// RemoveIndex removes the named index from the model definition. It returns an
// error if the index doesn't exist.
//
// This method should only be used to modify a model state during migration
// operations or to construct models programatically. Changing the model
// definition after it has been registered could cause unexpected errors.
func (m *Model) RemoveIndex(name string) error {
	if _, ok := m.meta.Indexes[name]; !ok {
		return fmt.Errorf("index not found: %s", name)
	}
	delete(m.meta.Indexes, name)
	return nil
}

// New creates a new model definition with the given arguments. It returns a
// Dispatcher embedding the model and holding the default Objects Manager.
//
// The model won't be ready to interact with the database until it's been
// registered to an application using either the Register function or the
// homonymous model method.
func New(name string, fields Fields, options Options) *Dispatcher {
	if options.Indexes == nil {
		options.Indexes = Indexes{}
	}
	model := &Model{name: name, fields: fields, meta: options}
	return &Dispatcher{
		model, Manager{Model: model, QuerySet: GenericQuerySet{}},
	}
}
