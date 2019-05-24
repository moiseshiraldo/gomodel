package gomodels

import (
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

func (i Instance) GetIf(field string) (Value, bool) {
	var val Value
	ok := true
	switch i.conType {
	case containers.Map:
		val, ok = i.Container.(Values)[field]
	case containers.Builder:
		val, ok = i.Container.(Builder).Get(field)
	default:
		cv := reflect.Indirect(reflect.ValueOf(i.Container))
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
