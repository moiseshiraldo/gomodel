package gomodel

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// The Container interface represents a type that holds the field values for
// a model object. It should be either a type implementing the Builder interface
// or a struct with the necessary exported fields.
type Container interface{}

// Getter is the interface that wraps the basic Get method for model containers.
//
// Get returns the value of the given field, and a boolean indicating whether
// the the field was found or not.
type Getter interface {
	Get(name string) (val Value, ok bool)
}

// Setter is the interface that wraps the basic Set method for model containers.
//
// Set receives the name, the value and the Field definition, and returns any
// error.
type Setter interface {
	Set(name string, val Value, field Field) error
}

// The Builder interface represents a model container that implements the Getter
// interface, the Setter interface and the New method to return a new empty
// container.
type Builder interface {
	Getter
	Setter
	New() Builder
}

// Value represents a value for a model field.
type Value interface{}

// Values is a map of field values that implements the Builder interface and
// is used as the default model Container.
type Values map[string]Value

// Get implements the Getter interface.
func (vals Values) Get(key string) (Value, bool) {
	val, ok := vals[key]
	return val, ok
}

// Set implements the Setter interface.
func (vals Values) Set(key string, val Value, field Field) error {
	recipient := field.Recipient()
	if err := setRecipient(recipient, val); err != nil {
		return err
	}
	vals[key] = reflect.Indirect(reflect.ValueOf(recipient)).Interface()
	return nil
}

// New implements the Builder interface.
func (vals Values) New() Builder {
	return Values{}
}

// Choice holds a choice option for a Field.
type Choice struct {
	Value Value  // Value is the choice value.
	Label string // Label is the choice label.
}

// isValidContainer checks if the given container is valid. It must be either
// a type implementing the Builder interface or a struct.
func isValidContainer(container Container) bool {
	if _, ok := container.(Builder); ok {
		return true
	}
	cv := reflect.Indirect(reflect.ValueOf(container))
	if cv.Kind() == reflect.Struct {
		return true
	}
	return false
}

// newContainer returns a new empty container of the same type as the given one.
func newContainer(container Container) Container {
	if b, ok := container.(Builder); ok {
		return b.New()
	}
	ct := reflect.Indirect(reflect.ValueOf(container)).Type()
	return reflect.New(ct).Interface()
}

// getRecipients returns a list of destination pointers for the given container
// and list of fields.
func getRecipients(con Container, fields []string, model *Model) []interface{} {
	recipients := make([]interface{}, 0, len(fields))
	if _, ok := con.(Setter); ok {
		for _, name := range fields {
			recipients = append(recipients, model.fields[name].Recipient())
		}
	} else {
		cv := reflect.Indirect(reflect.ValueOf(con))
		for _, name := range fields {
			f := cv.FieldByName(strings.Title(name))
			if f.IsValid() && f.CanSet() && f.CanAddr() {
				recipients = append(recipients, f.Addr().Interface())
			}
		}
	}
	return recipients
}

// getContainerField returns the container value for the field given by name
// and a boolean indicating whether the field was found or not.
func getContainerField(container Container, name string) (val Value, ok bool) {
	if getter, ok := container.(Getter); ok {
		if val, ok := getter.Get(name); ok {
			return val, true
		}
		return nil, false
	}
	cv := reflect.Indirect(reflect.ValueOf(container))
	field := cv.FieldByName(strings.Title(name))
	if field.IsValid() && field.CanInterface() {
		val := field.Interface()
		return val, true
	}
	return nil, false
}

// cloneBytes returns a copy of the given slice.
//
// Extracted from the database/sql package.
func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// asBytes returns the given value as a slice of bytes.
//
// Extracted from the database/sql package.
func asBytes(buf []byte, rv reflect.Value) (b []byte, ok bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return strconv.AppendUint(buf, rv.Uint(), 10), true
	case reflect.Float32:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 32), true
	case reflect.Float64:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 64), true
	case reflect.Bool:
		return strconv.AppendBool(buf, rv.Bool()), true
	case reflect.String:
		s := rv.String()
		return append(buf, s...), true
	}
	return
}

// asString returns the given value as a string
//
// Extracted from the database/sql package.
func asString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

// convertAssignRows copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
//
// Extracted from the database/sql package to mimic the Scan behaviour, with
// some variations.
func setRecipient(dest, src interface{}) error {
	if dest == nil {
		return fmt.Errorf("nil pointer recipient")
	}
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			*d = s
			return nil
		case *[]byte:
			*d = []byte(s)
			return nil
		case *sql.RawBytes:
			*d = append((*d)[:0], s...)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			*d = string(s)
			return nil
		case *interface{}:
			*d = cloneBytes(s)
			return nil
		case *[]byte:
			*d = cloneBytes(s)
			return nil
		case *sql.RawBytes:
			*d = s
			return nil
		}
	case time.Time:
		switch d := dest.(type) {
		case *time.Time:
			*d = s
			return nil
		case *string:
			*d = s.Format(time.RFC3339Nano)
			return nil
		case *[]byte:
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		case *sql.RawBytes:
			*d = s.AppendFormat((*d)[:0], time.RFC3339Nano)
			return nil
		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			*d = nil
			return nil
		case *[]byte:
			*d = nil
			return nil
		case *sql.RawBytes:
			*d = nil
			return nil
		}
	}

	var sv reflect.Value

	switch d := dest.(type) {
	case *string:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			*d = asString(src)
			return nil
		}
	case *[]byte:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes(nil, sv); ok {
			*d = b
			return nil
		}
	case *sql.RawBytes:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes([]byte(*d)[:0], sv); ok {
			*d = sql.RawBytes(b)
			return nil
		}
	case *bool:
		bv, err := driver.Bool.ConvertValue(src)
		if err == nil {
			*d = bv.(bool)
		}
		return err
	case *interface{}:
		*d = src
		return nil
	}

	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(src)
	}

	dpv := reflect.ValueOf(dest)

	if !sv.IsValid() {
		sv = reflect.ValueOf(src)
	}

	dv := reflect.Indirect(dpv)
	if sv.IsValid() && sv.Type().AssignableTo(dv.Type()) {
		switch b := src.(type) {
		case []byte:
			dv.Set(reflect.ValueOf(cloneBytes(b)))
		default:
			dv.Set(sv)
		}
		return nil
	}

	if dv.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dv.Type()) {
		dv.Set(sv.Convert(dv.Type()))
		return nil
	}

	switch dv.Kind() {
	case reflect.Ptr:
		if src == nil {
			dv.Set(reflect.Zero(dv.Type()))
			return nil
		}
		dv.Set(reflect.New(dv.Type().Elem()))
		return setRecipient(dv.Interface(), src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting type %T to a %s", src, dv.Kind())
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting type %T to a %s", src, dv.Kind())
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := asString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting type %T to a %s", src, dv.Kind())
		}
		dv.SetFloat(f64)
		return nil
	case reflect.String:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		switch v := src.(type) {
		case string:
			dv.SetString(v)
			return nil
		case []byte:
			dv.SetString(string(v))
			return nil
		}
	}

	return fmt.Errorf("error storing type %T into type %T", src, dest)
}
