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
	Get(f Filterer) (*Instance, error)
	Delete() (int64, error)
}

type GenericQuerySet struct {
	model       *Model
	constructor Constructor
	database    string
	columns     []string
	filter      Filterer
}

func (qs GenericQuerySet) dbError(err error) error {
	trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
	return &DatabaseError{qs.database, trace}
}

func (qs GenericQuerySet) constructorError(err error) error {
	trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
	return &ConstructorError{trace}
}

func (qs GenericQuerySet) addFilter(filter Filterer) GenericQuerySet {
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
	return qs.addFilter(filter)
}

func (qs GenericQuerySet) Load() ([]*Instance, error) {
	result := []*Instance{}
	consType := getConstructorType(qs.constructor)
	if consType == "" {
		return nil, qs.dbError(fmt.Errorf("invalid constructor type"))
	}
	db, ok := Databases[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	query, values := qs.Query()
	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, qs.dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		constructor, recipients := getRecipients(qs, consType)
		if len(recipients) != len(qs.columns) {
			err := fmt.Errorf("invalid constructor recipients")
			return nil, qs.constructorError(err)
		}
		err := rows.Scan(recipients...)
		if err != nil {
			return nil, qs.constructorError(err)
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
	err = rows.Err()
	if err != nil {
		return nil, qs.dbError(err)
	}
	return result, nil
}

func (qs GenericQuerySet) Get(filter Filterer) (*Instance, error) {
	qs = qs.addFilter(filter)
	consType := getConstructorType(qs.constructor)
	if consType == "" {
		return nil, qs.dbError(fmt.Errorf("invalid constructor type"))
	}
	db, ok := Databases[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	constructor, recipients := getRecipients(qs, consType)
	if len(recipients) != len(qs.columns) {
		err := fmt.Errorf("invalid constructor recipients")
		return nil, qs.constructorError(err)
	}
	query, values := qs.Query()
	err := db.QueryRow(query, values...).Scan(recipients...)
	if err != nil {
		return nil, qs.dbError(err)
	}
	if consType == "Map" {
		values := Values{}
		for i, name := range qs.columns {
			values[name] = reflect.ValueOf(recipients[i]).Elem()
		}
		constructor = values
	}
	instance := &Instance{constructor, qs.model}
	return instance, nil
}

func (qs GenericQuerySet) Delete() (int64, error) {
	db, ok := Databases[qs.database]
	if !ok {
		return 0, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	query := fmt.Sprintf("DELETE FROM %s", qs.model.Table())
	filter := ""
	values := make([]interface{}, 0)
	if qs.filter != nil {
		filter, values = qs.filter.Query()
		query += fmt.Sprintf(" WHERE %s", filter)
	}
	result, err := db.Exec(query, values...)
	if err != nil {
		return 0, qs.dbError(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, qs.dbError(err)
	}
	return count, nil
}
