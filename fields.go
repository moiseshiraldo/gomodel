package gomodel

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Field is the interface that represents a model field.
type Field interface {
	// IsPK returns true if the field is the model primary key.
	IsPK() bool
	// IsUnique returns true if the model field value must be unique.
	IsUnique() bool
	// IsNull returns true if the model field value can be null.
	IsNull() bool
	// IsAuto returns true if the field is an auto incremented one.
	IsAuto() bool
	// If IsAutoNow returns true, the field value will be time.Now() every
	// time a new row is inserted.
	IsAutoNow() bool
	// If IsAutoNowAdd returns true, the field value will be time.now() every
	// time a row is updated.
	IsAutoNowAdd() bool
	// If HasIndex returns true, the field column will be indexed.
	HasIndex() bool
	// DBColumn receives the model field name and returns the database column
	// name.
	DBColumn(fieldName string) string
	// DataType returns the field column type for the given driver.
	DataType(driver string) string
	// DefaultVal returns a default value for the field when hasDefault is true.
	DefaultVal() (val Value, hasDefault bool)
	// Recipient returns a pointer to the variable that will hold a field value
	// coming from the database.
	Recipient() interface{}
	// Value receives the recipient holding the field value for a particular
	// instance, and returns the Value that will be presented when the Instance
	// Get method is called.
	Value(recipient interface{}) Value
	// DriverValue receives a field value and returns the value that should
	// be used on database queries for the given driver.
	DriverValue(val Value, driver string) (interface{}, error)
	// DisplayValue returns the string representation of the given value.
	DisplayValue(val Value) string
}

// Fields represents the fields map of a model.
type Fields map[string]Field

// MarshalJSON implements the json.Marshaler interface.
func (fields Fields) MarshalJSON() ([]byte, error) {
	result := map[string]map[string]Field{}
	for name, f := range fields {
		m := map[string]Field{}
		m[strings.Split(reflect.ValueOf(f).Type().String(), ".")[1]] = f
		result[name] = m
	}
	return json.Marshal(result)
}

// UnmarshalJSON implements the json.Unmarshaler interface. An error is
// returned if the field type is not registered.
func (fields *Fields) UnmarshalJSON(data []byte) error {
	result := map[string]Field{}
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
			result[name] = fp.(Field)
		}
	}
	*fields = result
	return nil
}

// fieldsRegistry holds a global registry with the available fields.
var fieldsRegistry = Fields{
	"IntegerField": IntegerField{},
	"BooleanField": BooleanField{},
	"CharField":    CharField{},
	"DateField":    DateField{},
	"TimeField":    TimeField{},
}

// fieldInList returns true if name is found in the given list of fields.
func fieldInList(name string, fields []string) bool {
	for _, field := range fields {
		if field == name {
			return true
		}
	}
	return false
}
