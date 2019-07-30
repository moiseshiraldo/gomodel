package gomodels

import (
	"fmt"
)

type Manager struct {
	Model *Model
}

func (m Manager) Create(values Container) (*Instance, error) {
	db := databases["default"]
	container := m.Model.Container()
	instance := &Instance{m.Model, container}
	if !isValidContainer(values) {
		err := fmt.Errorf("invalid values container")
		return nil, &ContainerError{instance.trace(err)}
	}
	pk, err := db.InsertRow(m.Model, values)
	if err != nil {
		return instance, &DatabaseError{
			db.name,
			ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
		}
	}
	instance.Set(m.Model.pk, pk)
	for name, field := range m.Model.fields {
		var value Value
		if getter, ok := values.(Getter); ok {
			if val, ok := getter.Get(name); ok {
				value = val
			}
		} else if val, ok := getStructField(values, name); ok {
			value = val
		} else if val, hasDefault := field.DefaultVal(); hasDefault {
			value = val
		}
		if value != nil {
			if err := instance.Set(name, value); err != nil {
				return nil, err
			}
		}
	}
	return instance, nil
}

func (m Manager) GetQuerySet() QuerySet {
	cols := make([]string, 0, len(m.Model.fields))
	for name := range m.Model.fields {
		cols = append(cols, name)
	}
	return GenericQuerySet{
		model:     m.Model,
		container: m.Model.meta.Container,
		database:  "default",
		columns:   cols,
	}
}

func (m Manager) All() QuerySet {
	return m.GetQuerySet()
}

func (m Manager) Filter(c Conditioner) QuerySet {
	return m.GetQuerySet().Filter(c)
}

func (m Manager) Exclude(c Conditioner) QuerySet {
	return m.GetQuerySet().Exclude(c)
}

func (m Manager) Get(c Conditioner) (*Instance, error) {
	return m.GetQuerySet().Get(c)
}

func (m Manager) SetContainer(container Container) QuerySet {
	return m.GetQuerySet().SetContainer(container)
}
