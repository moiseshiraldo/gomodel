package gomodels

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Field interface {
	IsPK() bool
	IsUnique() bool
	IsNull() bool
	IsAuto() bool
	IsAutoNow() bool
	IsAutoNowAdd() bool
	HasIndex() bool
	DBColumn(fieldName string) string
	DataType(driver string) string
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
			field, ok := fieldsRegistry[fType]
			if !ok {
				return fmt.Errorf("invalid field type: %s", fType)
			}
			ft := reflect.Indirect(reflect.ValueOf(field)).Type()
			fp := reflect.New(ft).Interface()
			if err := json.Unmarshal(raw, fp); err != nil {
				return err
			}
			fields[name] = fp.(Field)
		}
	}
	*fp = fields
	return nil
}

var fieldsRegistry = Fields{
	"IntegerField": IntegerField{},
	"BooleanField": BooleanField{},
	"CharField":    CharField{},
	"DateField":    DateField{},
	"TimeField":    TimeField{},
}

func fieldInList(name string, fields []string) bool {
	for _, field := range fields {
		if field == name {
			return true
		}
	}
	return false
}
