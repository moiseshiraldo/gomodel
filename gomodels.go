package gomodels

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Field interface {
	IsPk() bool
	FromJSON(raw []byte) (Field, error)
}

type Fields map[string]Field

func (fields Fields) MarshalJSON() ([]byte, error) {
	result := map[string]map[string]Field{}
	for name, f := range fields {
		m := map[string]Field{}
		m[strings.Split(reflect.ValueOf(f).Type().String(), ".")[1]] = f
		result[name] = m
	}
	return json.Marshal(result)
}

func (fp *Fields) UnmarshalJSON(data []byte) error {
	fields := map[string]Field{}
	rawMap := map[string]map[string]json.RawMessage{}
	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}
	for name, fMap := range rawMap {
		for fType, raw := range fMap {
			native, ok := AvailableFields()[fType]
			if !ok {
				return fmt.Errorf("invalid field type: %s", fType)
			}
			field, err := native.FromJSON(raw)
			if err != nil {
				return err
			}
			fields[name] = field
		}
	}
	*fp = fields
	return nil
}

type Model struct {
	app    *Application
	name   string
	pk     string
	fields Fields
}

func (m Model) Name() string {
	return m.name
}

func (m Model) App() *Application {
	return m.app
}

func (m Model) Fields() Fields {
	return m.fields
}

func New(name string, fields Fields) *Model {
	return &Model{name: name, fields: fields}
}

func AvailableFields() map[string]Field {
	return map[string]Field{
		"IntegerField": IntegerField{},
		"AutoField":    AutoField{},
		"BooleanField": BooleanField{},
		"CharField":    CharField{},
	}
}
