package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

type QuerySet interface {
	Load() ([]*Instance, error)
	Query() (placeholder string, values []interface{})
	Model() *Model
	Columns() []string
	Constructor() Constructor
	Filter(f Filterer) QuerySet
}

type GenericQuerySet struct {
	model       *Model
	constructor Constructor
	database    string
	columns     []string
	filter      Filterer
}

func (qs GenericQuerySet) Query() (string, []interface{}) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s", strings.Join(qs.columns, ", "), qs.model.Table(),
	)
	if qs.filter != nil {
		filter, values := qs.filter.Query()
		query += fmt.Sprintf(" WHERE %s", filter)
		return query, values
	} else {
		return query, make([]interface{}, 0)
	}
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

func (qs GenericQuerySet) Filter(filter Filterer) QuerySet {
	if qs.filter == nil {
		if query, ok := filter.(Q); ok {
			qs.filter = Filter{sibs: []Filterer{query}}
		} else {
			qs.filter = filter
		}
	} else {
		qs.filter = qs.filter.And(filter)
	}
	return qs
}

func (qs GenericQuerySet) Load() ([]*Instance, error) {
	result := []*Instance{}
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
	query, values := qs.Query()
	rows, err := db.Query(query, values...)
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
		result = append(result, &Instance{constructor, qs.model})
	}
	return result, nil
}
