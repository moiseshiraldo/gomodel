package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

type QuerySet interface {
	Load() ([]*Instance, error)
	Query() string
}

type GenericQuerySet struct {
	model       *Model
	constructor Constructor
	database    string
	columns     []string
}

func (qs GenericQuerySet) Query() string {
	return fmt.Sprintf(
		"SELECT %s FROM %s", strings.Join(qs.columns, ", "), qs.model.Table(),
	)
}

func (qs GenericQuerySet) Load() ([]*Instance, error) {
	result := []*Instance{}
	db := Databases[qs.database]
	rows, err := db.Query(qs.Query())
	if err != nil {
		trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
		return nil, &DatabaseError{qs.database, trace}
	}
	defer rows.Close()
	for rows.Next() {
		var constructor Constructor
		recipients := make([]interface{}, 0, len(qs.columns))
		switch qs.constructor.(type) {
		case Values:
			constructor = Values{}
			for _, name := range qs.columns {
				val := qs.model.fields[name].NativeVal()
				recipients = append(recipients, &val)
			}
		default:
			if builder, ok := qs.constructor.(Builder); ok {
				builder = builder.New()
				recipients = builder.Recipients(qs.columns)
				constructor = builder
			} else {
				ct := reflect.TypeOf(qs.constructor)
				if ct.Kind() == reflect.Ptr {
					ct = ct.Elem()
				}
				cp := reflect.New(ct)
				for _, name := range qs.columns {
					f := cp.Elem().FieldByName(strings.Title(name))
					if !f.IsValid() || !f.CanAddr() {
						return result, fmt.Errorf("field not found %s", name)
					}
					recipients = append(recipients, f.Addr().Interface())
				}
				constructor = cp.Interface()
			}
		}
		err := rows.Scan(recipients...)
		if err != nil {
			trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
			return nil, &DatabaseError{qs.database, trace}
		}
		if _, ok := qs.constructor.(Values); ok {
			values := Values{}
			for i, name := range qs.columns {
				values[name] = reflect.ValueOf(recipients[i]).Elem()
			}
			constructor = values
		}
		result = append(result, &Instance{constructor, qs.model})
	}
	return result, nil
}
