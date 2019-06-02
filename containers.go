package gomodels

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
)

var containers = struct {
	Map     string
	Builder string
	Struct  string
}{"Map", "Builder", "Struct"}

type Value interface{}
type Values map[string]Value

type Container interface{}

type Builder interface {
	Get(field string) (val Value, ok bool)
	Set(field string, val Value) error
	New() Builder
	Recipients(columns []string) []interface{}
}

func (vals Values) Get(field string) (Value, bool) {
	val, ok := vals[field]
	return val, ok
}

func (vals Values) Set(field string, val Value) bool {
	vals[field] = val
	return true
}

func (vals Values) New() Container {
	return Values{}
}

func (vals Values) Recipients(columns []string) []interface{} {
	return nil
}

type Instance struct {
	Container
	conType string
	Model   *Model
}

func (i Instance) trace(err error) ErrorTrace {
	return ErrorTrace{App: i.Model.app, Model: i.Model, Err: err}
}

func (i Instance) GetIf(field string) (Value, bool) {
	switch i.conType {
	case containers.Map:
		val, ok := i.Container.(Values)[field]
		if ok {
			if vlr, isVlr := val.(driver.Valuer); isVlr {
				if val, err := vlr.Value(); err == nil {
					return val, true
				}
			}
		}
		return val, ok
	case containers.Builder:
		val, ok := i.Container.(Builder).Get(field)
		return val, ok
	default:
		cv := reflect.Indirect(reflect.ValueOf(i.Container))
		f := cv.FieldByName(strings.Title(field))
		if f.IsValid() && f.CanInterface() {
			val := f.Interface()
			if vlr, isVlr := val.(driver.Valuer); isVlr {
				if val, err := vlr.Value(); err == nil {
					return val, true
				}
			}
			return val, true
		} else {
			return nil, false
		}
	}
}

func (i Instance) Get(field string) Value {
	val, _ := i.GetIf(field)
	return val
}

func (i Instance) Set(field string, val Value) error {
	switch i.conType {
	case containers.Map:
		i.Container.(Values)[field] = val
		return nil
	case containers.Builder:
		return i.Container.(Builder).Set(field, val)
	default:
		cv := reflect.Indirect(reflect.ValueOf(i.Container))
		f := cv.FieldByName(strings.Title(field))
		if !f.IsValid() || !f.CanSet() {
			return fmt.Errorf("Invalid field")
		}
		f.Set(reflect.ValueOf(val))
		return nil
	}
}

func (i Instance) Save(fields ...string) error {
	pkVal, hasPk := i.GetIf(i.Model.pk)
	if !hasPk {
		return &ContainerError{i.trace(fmt.Errorf("container missing pk"))}
	}
	db := Databases["default"]
	_, autoPk := i.Model.fields[i.Model.pk].(AutoField)
	if autoPk && pkVal == reflect.Zero(reflect.TypeOf(pkVal)).Interface() {
		query, vals := sqlInsertQuery(i, fields)
		result, err := db.Exec(query, vals...)
		if err != nil {
			return &DatabaseError{"default", i.trace(err)}
		}
		id, err := result.LastInsertId()
		if err != nil {
			return &DatabaseError{"default", i.trace(err)}
		}
		i.Set(i.Model.pk, id)
	} else {
		query, vals := sqlUpdateQuery(i, fields)
		result, err := db.Exec(query, vals...)
		if err != nil {
			return &DatabaseError{"default", i.trace(err)}
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return &DatabaseError{"default", i.trace(err)}
		}
		if rows == 0 {
			query, vals := sqlInsertQuery(i, fields)
			_, err := db.Exec(query, vals...)
			if err != nil {
				return &DatabaseError{"default", i.trace(err)}
			}
		}
	}
	return nil
}
