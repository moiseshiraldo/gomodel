package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

type Value interface{}
type Values map[string]Value

type Constructor interface{}

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

func (vals Values) New() Constructor {
	return Values{}
}

func (vals Values) Recipients(columns []string) []interface{} {
	return nil
}

type Instance struct {
	Constructor
	Model *Model
}

func (i Instance) Get(field string) Value {
	var val Value
	switch constructor := i.Constructor.(type) {
	case Values:
		val = constructor[field]
	case Builder:
		val, _ = constructor.Get(field)
	default:
		cv := reflect.Indirect(reflect.ValueOf(constructor))
		val = cv.FieldByName(strings.Title(field)).Interface()
	}
	return val
}

func (i Instance) Set(field string, val Value) error {
	switch constructor := i.Constructor.(type) {
	case Values:
		constructor[field] = val
		return nil
	case Builder:
		return constructor.Set(field, val)
	default:
		r := reflect.ValueOf(i.Constructor)
		f := r.Elem().FieldByName("N")
		if !f.IsValid() || !f.CanSet() {
			return fmt.Errorf("Invalid field")
		}
		f.Set(reflect.ValueOf(val))
		return nil
	}
}
