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

func (i Instance) GetIf(field string) (Value, bool) {
	var val Value
	ok := true
	switch constructor := i.Constructor.(type) {
	case Values:
		val = constructor[field]
	case Builder:
		val, ok = constructor.Get(field)
	default:
		cv := reflect.Indirect(reflect.ValueOf(constructor))
		f := cv.FieldByName(strings.Title(field))
		if f.IsValid() && f.CanInterface() {
			val = f.Interface()
		} else {
			ok = false
		}
	}
	return val, ok
}

func (i Instance) Get(field string) Value {
	val, _ := i.GetIf(field)
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
		cv := reflect.Indirect(reflect.ValueOf(constructor))
		f := cv.FieldByName(strings.Title(field))
		if !f.IsValid() || !f.CanSet() {
			return fmt.Errorf("Invalid field")
		}
		f.Set(reflect.ValueOf(val))
		return nil
	}
}
