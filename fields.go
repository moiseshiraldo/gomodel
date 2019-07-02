package gomodels

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Field interface {
	IsPk() bool
	IsAuto() bool
	DBColumn(fieldName string) string
	HasIndex() bool
	SqlDatatype(driver string) string
	DefaultVal() (val Value, hasDefault bool)
	Recipient() interface{}
	Value(recipient interface{}) Value
	DriverValue(val Value, driver string) (interface{}, error)
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
			field, ok := AvailableFields()[fType]
			if !ok {
				return fmt.Errorf("invalid field type: %s", fType)
			}
			if err := json.Unmarshal(raw, &field); err != nil {
				fmt.Println(err)
				return err
			}
			fields[name] = field
		}
	}
	*fp = fields
	return nil
}

func AvailableFields() Fields {
	return Fields{
		"IntegerField": &IntegerField{},
		"AutoField":    &AutoField{},
		"BooleanField": &BooleanField{},
		"CharField":    &CharField{},
		"DateField":    &DateField{},
	}
}

func fieldInList(name string, fields []string) bool {
	for _, field := range fields {
		if field == name {
			return true
		}
	}
	return false
}

func sqlColumnOptions(null bool, pk bool, unique bool) string {
	options := ""
	if null {
		options += " NULL"
	} else {
		options += " NOT NULL"
	}
	if pk {
		options += " PRIMARY KEY"
	} else if unique {
		options += " UNIQUE"
	}
	return options
}
