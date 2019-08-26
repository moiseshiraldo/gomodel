package gomodels

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Instance struct {
	model     *Model
	container Container
}

func (i Instance) trace(err error) ErrorTrace {
	return ErrorTrace{App: i.model.app, Model: i.model, Err: err}
}

func (i Instance) Container() Container {
	return i.container
}

func (i Instance) Model() *Model {
	return i.model
}

func (i Instance) GetIf(key string) (Value, bool) {
	field, ok := i.model.fields[key]
	if !ok {
		return nil, false
	}
	val, ok := getContainerField(i.container, key)
	if !ok {
		return nil, false
	}
	return field.Value(val), true
}

func (i Instance) Get(key string) Value {
	val, _ := i.GetIf(key)
	return val
}

func (i Instance) Set(key string, val Value) error {
	field, ok := i.model.fields[key]
	if !ok {
		return &ContainerError{i.trace(fmt.Errorf("unknown field %s", key))}
	}
	if c, ok := i.container.(Setter); ok {
		if err := c.Set(key, val, field); err != nil {
			return &ContainerError{i.trace(err)}
		}
	} else {
		cv := reflect.Indirect(reflect.ValueOf(i.container))
		f := cv.FieldByName(strings.Title(key))
		if !f.IsValid() || !f.CanSet() || !f.CanAddr() {
			return &ContainerError{i.trace(fmt.Errorf("Invalid field"))}
		}
		if err := setRecipient(f.Addr().Interface(), val); err != nil {
			return &ContainerError{i.trace(err)}
		}
	}
	return nil
}

func (i Instance) SetValues(values Container) error {
	for name := range i.model.fields {
		if val, ok := getContainerField(values, name); ok {
			if err := i.Set(name, val); err != nil {
				return &ContainerError{i.trace(err)}
			}
		}
	}
	return nil
}

func (i Instance) valueToSave(name string, creating bool) (Value, error) {
	field, ok := i.model.fields[name]
	if !ok {
		err := fmt.Errorf("unknown field: %s", name)
		return nil, err
	}
	if field.IsAuto() {
		return nil, nil
	}
	if field.IsAutoNow() || creating && field.IsAutoNowAdd() {
		val := time.Now()
		if err := i.Set(name, val); err != nil {
			return nil, err
		}
		return val, nil
	}
	if val, ok := getContainerField(i.container, name); ok {
		return val, nil
	} else if val, hasDefault := field.DefaultVal(); creating && hasDefault {
		if err := i.Set(name, val); err != nil {
			return nil, err
		}
		return val, nil
	}
	return nil, nil
}

func (i Instance) insertRow(db Database, autoPk bool, fields ...string) error {
	dbValues := Values{}
	for _, name := range fields {
		val, err := i.valueToSave(name, true)
		if err != nil {
			return &ContainerError{i.trace(err)}
		} else if val != nil {
			dbValues[name] = val
		}
	}
	pk, err := db.InsertRow(i.model, dbValues)
	if err != nil {
		return &DatabaseError{db.id, i.trace(err)}
	}
	if autoPk {
		if err := i.Set(i.model.pk, pk); err != nil {
			return err
		}
	}
	return nil
}

func (i Instance) updateRow(db Database, pkVal Value, fields ...string) error {
	dbValues := Values{}
	for _, name := range fields {
		if name == i.model.pk {
			continue
		}
		val, err := i.valueToSave(name, false)
		if err != nil {
			return &ContainerError{i.trace(err)}
		} else if val != nil {
			dbValues[name] = val
		}
	}
	rows, err := db.UpdateRows(i.model, dbValues, Q{i.model.pk: pkVal})
	if err != nil {
		return &DatabaseError{db.id, i.trace(err)}
	}
	if rows == 0 {
		return i.insertRow(db, false, fields...)
	}
	return nil
}

func (i Instance) Save(fields ...string) error {
	db := dbRegistry["default"]
	if len(fields) == 0 {
		for name := range i.model.fields {
			fields = append(fields, name)
		}
	}
	autoPk := i.model.fields[i.model.pk].IsAuto()
	pkVal := i.Get(i.model.pk)
	if pkVal != nil {
		zero := reflect.Zero(reflect.TypeOf(pkVal)).Interface()
		if !(autoPk && pkVal == zero) {
			return i.updateRow(db, pkVal, fields...)
		}
	}
	return i.insertRow(db, autoPk, fields...)
}
