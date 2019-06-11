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
	if c, ok := i.container.(Getter); ok {
		return c.Get(key)
	} else {
		cv := reflect.Indirect(reflect.ValueOf(i.container))
		f := cv.FieldByName(strings.Title(key))
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

func (i Instance) Get(key string) Value {
	val, _ := i.GetIf(key)
	return val
}

func (i Instance) Set(key string, val Value) error {
	field, _ := i.model.fields[key]
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