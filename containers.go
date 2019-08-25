package gomodels

import (
	"reflect"
	"strings"
)

var containers = struct {
	Map     string
	Builder string
	Struct  string
}{"Map", "Builder", "Struct"}

type Container interface{}

type Getter interface {
	Get(key string) (val Value, ok bool)
}

type Setter interface {
	Set(key string, val Value, field Field) error
}

type Builder interface {
	Getter
	Setter
	New() Builder
}

type Value interface{}
type Values map[string]Value

func (vals Values) Get(key string) (Value, bool) {
	val, ok := vals[key]
	return val, ok
}

func (vals Values) Set(key string, val Value, field Field) error {
	recipient := field.Recipient()
	if err := setRecipient(recipient, val); err != nil {
		return err
	}
	vals[key] = reflect.Indirect(reflect.ValueOf(recipient)).Interface()
	return nil
}

func (vals Values) New() Builder {
	return Values{}
}

func isValidContainer(container Container) bool {
	if _, ok := container.(Builder); ok {
		return true
	} else {
		cv := reflect.Indirect(reflect.ValueOf(container))
		if cv.Kind() == reflect.Struct {
			return true
		}
	}
	return false
}

func getRecipients(con Container, cols []string, model *Model) []interface{} {
	recipients := make([]interface{}, 0, len(cols))
	if _, ok := con.(Setter); ok {
		for _, name := range cols {
			recipients = append(recipients, model.fields[name].Recipient())
		}
	} else {
		cv := reflect.Indirect(reflect.ValueOf(con))
		for _, name := range cols {
			f := cv.FieldByName(strings.Title(name))
			if f.IsValid() && f.CanSet() && f.CanAddr() {
				recipients = append(recipients, f.Addr().Interface())
			}
		}
	}
	return recipients
}

func getContainerField(container Container, name string) (Value, bool) {
	if getter, ok := container.(Getter); ok {
		if val, ok := getter.Get(name); ok {
			return val, true
		}
	} else {
		cv := reflect.Indirect(reflect.ValueOf(container))
		field := cv.FieldByName(strings.Title(name))
		if field.IsValid() && field.CanInterface() {
			val := field.Interface()
			return val, true
		} else {
			return nil, false
		}
	}
	return nil, false
}
