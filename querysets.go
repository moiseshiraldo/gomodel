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
		recipients := make([]interface{}, 0, len(qs.columns))
		for _, name := range qs.columns {
			val := qs.model.fields[name].NativeVal()
			recipients = append(recipients, &val)
		}
		err := rows.Scan(recipients...)
		if err != nil {
			trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
			return nil, &DatabaseError{qs.database, trace}
		}
		constructor := qs.constructor.New()
		for i, name := range qs.columns {
			constructor.Set(name, reflect.ValueOf(recipients[i]).Elem())
		}
		ins := &Instance{constructor, qs.model}
		result = append(result, ins)
	}
	return result, nil
}
