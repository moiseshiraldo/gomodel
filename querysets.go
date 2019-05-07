package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

type QuerySet interface {
	Load() ([]Instance, error)
	Query() string
	Model() *Model
	Columns() []string
	Constructor() Constructor
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

func (qs GenericQuerySet) Model() *Model {
	return qs.model
}

func (qs GenericQuerySet) Columns() []string {
	return qs.columns
}

func (qs GenericQuerySet) Constructor() Constructor {
	return qs.constructor
}

func (qs GenericQuerySet) Load() ([]Instance, error) {
	result := []Instance{}
	trace := ErrorTrace{App: qs.model.app, Model: qs.model}
	consType := getConstructorType(qs.constructor)
	if consType == "" {
		trace.Err = fmt.Errorf("invalid constructor type")
		return nil, &ConstructorError{trace}
	}
	db, ok := Databases[qs.database]
	if !ok {
		trace.Err = fmt.Errorf("db not found: %s", qs.database)
		return nil, &DatabaseError{qs.database, trace}
	}
	rows, err := db.Query(qs.Query())
	if err != nil {
		trace.Err = err
		return nil, &DatabaseError{qs.database, trace}
	}
	defer rows.Close()
	for rows.Next() {
		constructor, recipients := getRecipients(qs, consType)
		if len(recipients) != len(qs.columns) {
			trace.Err = fmt.Errorf("invalid constructor recipients")
			return nil, &ConstructorError{trace}
		}
		err := rows.Scan(recipients...)
		if err != nil {
			trace.Err = err
			return nil, &ConstructorError{trace}
		}
		if consType == "Map" {
			values := Values{}
			for i, name := range qs.columns {
				values[name] = reflect.ValueOf(recipients[i]).Elem()
			}
			constructor = values
		}
		result = append(result, Instance{constructor, qs.model})
	}
	return result, nil
}
