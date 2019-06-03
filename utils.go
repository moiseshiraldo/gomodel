package gomodels

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

func sqlCreateQuery(table string, values Values) (string, []interface{}) {
	cols := make([]string, 0, len(values))
	vals := make([]interface{}, 0, len(values))
	placeholders := make([]string, 0, len(values))
	index := 1
	for col, val := range values {
		cols = append(cols, fmt.Sprintf("'%s'", col))
		vals = append(vals, val)
		placeholders = append(placeholders, fmt.Sprintf("$%d", index))
		index += 1
	}
	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "),
	)
	return query, vals
}

func sqlInsertQuery(i Instance, fields []string) (string, []interface{}) {
	vals := make([]interface{}, 0, len(i.Model.fields))
	cols := make([]string, 0, len(i.Model.fields))
	placeholders := make([]string, 0, len(i.Model.fields))
	if len(fields) == 0 {
		index := 1
		for name := range i.Model.fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s'", name))
				vals = append(vals, val)
				placeholders = append(placeholders, fmt.Sprintf("$%d", index))
			}
			index += 1
		}
	} else {
		for index, name := range fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s'", name))
				vals = append(vals, val)
				placeholders = append(
					placeholders, fmt.Sprintf("$%d", index+1),
				)
			}
		}
	}
	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		i.Model.Table(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	return query, vals
}

func sqlUpdateQuery(i Instance, fields []string) (string, []interface{}) {
	vals := make([]interface{}, 0, len(i.Model.fields))
	cols := make([]string, 0, len(i.Model.fields))
	if len(fields) == 0 {
		index := 1
		for name := range i.Model.fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s' = $%d", name, index))
				vals = append(vals, val)
			}
			index += 1
		}
	} else {
		for index, name := range fields {
			if name == i.Model.pk {
				continue
			}
			if val, ok := i.GetIf(name); ok {
				cols = append(cols, fmt.Sprintf("'%s' = $%d", name, index+1))
				vals = append(vals, val)
			}
		}
	}
	vals = append(vals, i.Get(i.Model.pk))
	query := fmt.Sprintf(
		"UPDATE \"%s\" SET %s WHERE \"%s\" = $%d",
		i.Model.Table(),
		strings.Join(cols, ", "),
		i.Model.pk,
		len(cols)+1,
	)
	return query, vals
}

func getContainerType(container Container) (string, error) {
	switch container.(type) {
	case Values:
		return containers.Map, nil
	default:
		if _, ok := container.(Builder); ok {
			return containers.Builder, nil
		} else {
			ct := reflect.TypeOf(container)
			if ct.Kind() == reflect.Ptr {
				ct = ct.Elem()
			}
			if ct.Kind() == reflect.Struct {
				return containers.Struct, nil
			}
		}
		return "", fmt.Errorf("invlid container")
	}
}

func getRecipients(qs QuerySet, conType string) (Container, []interface{}) {
	container := qs.Container()
	recipients := make([]interface{}, 0, len(qs.Columns()))
	switch conType {
	case containers.Map:
		for _, name := range qs.Columns() {
			recipients = append(
				recipients, qs.Model().fields[name].Recipient(),
			)
		}
	case containers.Builder:
		recipients = container.(Builder).Recipients(qs.Columns())
	default:
		cv := reflect.Indirect(reflect.ValueOf(container))
		for _, name := range qs.Columns() {
			f := cv.FieldByName(strings.Title(name))
			if f.IsValid() && f.CanAddr() {
				recipients = append(recipients, f.Addr().Interface())
			}
		}
	}
	return container, recipients
}

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

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

func setContainerField(dest, src interface{}) error {
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = []byte(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = append((*d)[:0], s...)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = string(s)
			return nil
		case *interface{}:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = cloneBytes(s)
			return nil
		case *[]byte:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = cloneBytes(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
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
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		case *sql.RawBytes:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = s.AppendFormat((*d)[:0], time.RFC3339Nano)
			return nil
		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
			*d = nil
			return nil
		case *sql.RawBytes:
			if d == nil {
				return fmt.Errorf("nil pointer")
			}
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
		return setContainerField(dv.Interface(), src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf(
				"converting driver.Value type %T to a %s", src, dv.Kind(),
			)
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
			return fmt.Errorf(
				"converting driver.Value type %T to a %s", src, dv.Kind(),
			)
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
			return fmt.Errorf(
				"converting driver.Value type %T to a %s", src, dv.Kind(),
			)
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
