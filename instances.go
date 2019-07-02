package gomodels

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
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
	var val Value
	var exists bool
	if c, ok := i.container.(Getter); ok {
		val, exists = c.Get(key)
	} else {
		val, exists = getStructField(i.container, key)
	}
	if field, ok := i.model.fields[key]; ok && exists {
		return field.Value(val), true
	}
	return nil, false
}

func (i Instance) Get(key string) Value {
	val, _ := i.GetIf(key)
	return val
}

func (i Instance) Set(key string, val Value) error {
	field, _ := i.model.fields[key]
	if vlr, isVlr := val.(driver.Valuer); isVlr {
		if v, err := vlr.Value(); err == nil {
			val = v
		}
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
		if err := setContainerField(f.Addr().Interface(), val); err != nil {
			return &ContainerError{i.trace(err)}
		}
	}
	return nil
}

func (i Instance) Save(fields ...string) error {
	pkVal, hasPk := i.GetIf(i.model.pk)
	if !hasPk {
		return &ContainerError{i.trace(fmt.Errorf("container missing pk"))}
	}
	db := databases["default"]
	autoPk := i.model.fields[i.model.pk].IsAuto()
	if autoPk && pkVal == reflect.Zero(reflect.TypeOf(pkVal)).Interface() {
		pk, err := db.InsertRow(i.model, i.container, fields...)
		if err != nil {
			return &DatabaseError{db.name, i.trace(err)}
		}
		if err := i.Set(i.model.pk, pk); err != nil {
			return err
		}
	} else {
		rows, err := db.UpdateRows(
			i.model, i.container, Q{i.model.pk: pkVal}, fields...,
		)
		if err != nil {
			return &DatabaseError{db.name, i.trace(err)}
		}
		if rows == 0 {
			_, err := db.InsertRow(i.model, i.container, fields...)
			if err != nil {
				return &DatabaseError{db.name, i.trace(err)}
			}
		}
	}
	return nil
}
