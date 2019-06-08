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
	model     *Model
	container Container
	conType   string
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

func (i Instance) GetIf(field string) (Value, bool) {
	switch i.conType {
	case containers.Map:
		val, ok := i.container.(Values)[field]
		if ok {
			if vlr, isVlr := val.(driver.Valuer); isVlr {
				if val, err := vlr.Value(); err == nil {
					return val, true
				}
			}
		}
		return val, ok
	case containers.Builder:
		val, ok := i.container.(Builder).Get(field)
		return val, ok
	default:
		cv := reflect.Indirect(reflect.ValueOf(i.container))
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
		if f, ok := i.model.fields[field]; ok {
			recipient := f.Recipient()
			if err := setContainerField(recipient, val); err != nil {
				return &ContainerError{i.trace(err)}
			}
			i.container.(Values)[field] = reflect.Indirect(
				reflect.ValueOf(recipient),
			).Interface()
		} else {
			i.container.(Values)[field] = val
		}
	case containers.Builder:
		if err := i.container.(Builder).Set(field, val); err != nil {
			return &ContainerError{i.trace(err)}
		}
	default:
		cv := reflect.Indirect(reflect.ValueOf(i.container))
		f := cv.FieldByName(strings.Title(field))
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
	db := Databases["default"]
	_, autoPk := i.model.fields[i.model.pk].(AutoField)
	if autoPk && pkVal == reflect.Zero(reflect.TypeOf(pkVal)).Interface() {
		query, vals := sqlInsertQuery(i, fields, db.Driver)
		if db.Driver == "postgres" {
			var pk int64
			err := db.conn.QueryRow(query, vals...).Scan(&pk)
			if err != nil {
				return &DatabaseError{"default", i.trace(err)}
			}
			i.Set(i.model.pk, pk)
		} else {
			result, err := db.conn.Exec(query, vals...)
			if err != nil {
				return &DatabaseError{"default", i.trace(err)}
			}
			id, err := result.LastInsertId()
			if err != nil {
				return &DatabaseError{"default", i.trace(err)}
			}
			i.Set(i.model.pk, id)
		}
	} else {
		query, vals := sqlUpdateQuery(i, fields)
		result, err := db.conn.Exec(query, vals...)
		if err != nil {
			return &DatabaseError{"default", i.trace(err)}
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return &DatabaseError{"default", i.trace(err)}
		}
		if rows == 0 {
			query, vals := sqlInsertQuery(i, fields, db.Driver)
			_, err := db.conn.Exec(query, vals...)
			if err != nil {
				return &DatabaseError{"default", i.trace(err)}
			}
		}
	}
	return nil
}
