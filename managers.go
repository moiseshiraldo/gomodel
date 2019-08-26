package gomodels

import (
	"fmt"
	"time"
)

type Manager struct {
	Model *Model
}

func (m Manager) Create(values Container) (*Instance, error) {
	db := dbRegistry["default"]
	container := m.Model.Container()
	instance := &Instance{m.Model, container}
	if !isValidContainer(values) {
		err := fmt.Errorf("invalid values container")
		return nil, &ContainerError{instance.trace(err)}
	}
	dbValues := Values{}
	for name, field := range m.Model.fields {
		if field.IsAuto() {
			continue
		}
		var dbVal Value
		if field.IsAutoNowAdd() {
			dbVal = time.Now()
		}
		if val, ok := getContainerField(values, name); ok {
			dbVal = val
		} else if val, hasDefault := field.DefaultVal(); hasDefault {
			dbVal = val
		}
		if dbVal != nil {
			dbValues[name] = dbVal
			if err := instance.Set(name, dbVal); err != nil {
				return nil, err
			}
		}
	}
	pk, err := db.InsertRow(m.Model, dbValues)
	if err != nil {
		trace := ErrorTrace{App: m.Model.app, Model: m.Model, Err: err}
		return instance, &DatabaseError{db.id, trace}
	}
	if err := instance.Set(m.Model.pk, pk); err != nil {
		return nil, err
	}
	return instance, nil
}

func (m Manager) GetQuerySet() QuerySet {
	fields := make([]string, 0, len(m.Model.fields))
	for name := range m.Model.fields {
		fields = append(fields, name)
	}
	return GenericQuerySet{
		model:     m.Model,
		container: m.Model.meta.Container,
		database:  "default",
		fields:    fields,
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
