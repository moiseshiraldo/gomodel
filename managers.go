package gomodel

import (
	"fmt"
	"time"
)

// A Manager is the interface through which database query operations are
// provided to models.
type Manager struct {
	// Model is the model making the queries.
	Model *Model
	// QuerySet is the base queryset for the manager.
	QuerySet QuerySet
	tx       *Transaction
}

// WithTx returns a new manager where all database query operations will be
// applied on the given transaction.
func (m Manager) WithTx(tx *Transaction) Manager {
	m.tx = tx
	return m
}

// Create makes a new object with the given values, saves it to the default
// database and returns the instance representing the object.
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

// GetQuerySet calls the New method of the base QuerySet and returns the result.
func (m Manager) GetQuerySet() QuerySet {
	return m.QuerySet.New(m.Model, m.QuerySet)
}

// All returns a QuerySet representing all objects.
func (m Manager) All() QuerySet {
	return m.GetQuerySet()
}

// Filter returns a QuerySet filtered by the given conditioner.
func (m Manager) Filter(c Conditioner) QuerySet {
	return m.GetQuerySet().Filter(c)
}

// Exclude returns a QuerySet exluding objects by the given conditioner.
func (m Manager) Exclude(c Conditioner) QuerySet {
	return m.GetQuerySet().Exclude(c)
}

// Get returns an instance representing the single object matching the given
// conditioner.
//
// If no object is found, *ObjectNotFoundError is returned.
//
// If multiple objects match the conditions, *MultipleObjectsError is returned.
func (m Manager) Get(c Conditioner) (*Instance, error) {
	return m.GetQuerySet().Get(c)
}

// WithContainer returns a QuerySet with the given Container type as a base.
func (m Manager) WithContainer(container Container) QuerySet {
	return m.GetQuerySet().WithContainer(container)
}
