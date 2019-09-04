package gomodel

import (
	"fmt"
	"time"
)

type Manager struct {
	Model    *Model
	QuerySet QuerySet
	tx       *Transaction
}

func (m Manager) WithTx(tx *Transaction) Manager {
	m.tx = tx
	return m
}

func (m Manager) Create(values Container) (*Instance, error) {
	db := dbRegistry["default"]
	engine := db.Engine
	if m.tx != nil {
		engine = m.tx.Engine
	}
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
	pk, err := engine.InsertRow(m.Model, dbValues)
	if err != nil {
		if m.tx != nil {
			db = m.tx.DB
		}
		trace := ErrorTrace{App: m.Model.app, Model: m.Model, Err: err}
		return instance, &DatabaseError{db.id, trace}
	}
	if m.Model.fields[m.Model.pk].IsAuto() {
		if err := instance.Set(m.Model.pk, pk); err != nil {
			return nil, err
		}
	}
	return instance, nil
}

func (m Manager) GetQuerySet() QuerySet {
	return m.QuerySet.New(m.Model, m.QuerySet)
}

func (m Manager) All() QuerySet {
	return m.GetQuerySet()
}

func (m Manager) Filter(c Conditioner) QuerySet {
	return m.All().Filter(c)
}

func (m Manager) Exclude(c Conditioner) QuerySet {
	return m.All().Exclude(c)
}

func (m Manager) Get(c Conditioner) (*Instance, error) {
	return m.All().Get(c)
}

func (m Manager) WithContainer(container Container) QuerySet {
	return m.All().WithContainer(container)
}
