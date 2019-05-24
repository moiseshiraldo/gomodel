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
	Container() Container
	Filter(f Filterer) QuerySet
	Get(f Filterer) (*Instance, error)
	Delete() (int64, error)
}

type GenericQuerySet struct {
	model     *Model
	container Container
	conType   string
	database  string
	columns   []string
	filter    Filterer
}

func (qs GenericQuerySet) dbError(err error) error {
	trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
	return &DatabaseError{qs.database, trace}
}

func (qs GenericQuerySet) containerError(err error) error {
	trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
	return &ContainerError{trace}
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

func (qs GenericQuerySet) Container() Container {
	return qs.container
}

func (qs GenericQuerySet) Filter(filter Filterer) QuerySet {
	return qs.addFilter(filter)
}

func (qs GenericQuerySet) Load() ([]*Instance, error) {
	result := []*Instance{}
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
		container, recipients := getRecipients(qs, qs.conType)
		if len(recipients) != len(qs.columns) {
			err := fmt.Errorf("invalid container recipients")
			return nil, qs.containerError(err)
		}
		err := rows.Scan(recipients...)
		if err != nil {
			return nil, qs.containerError(err)
		}
		if qs.conType == containers.Map {
			values := Values{}
			for i, name := range qs.columns {
				values[name] = reflect.ValueOf(recipients[i]).Elem()
			}
			container = values
		}
		result = append(result, &Instance{container, qs.conType, qs.model})
	}
	err = rows.Err()
	if err != nil {
		return nil, qs.dbError(err)
	}
	return result, nil
}

func (qs GenericQuerySet) Get(filter Filterer) (*Instance, error) {
	qs = qs.addFilter(filter)
	db, ok := Databases[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	container, recipients := getRecipients(qs, qs.conType)
	if len(recipients) != len(qs.columns) {
		err := fmt.Errorf("invalid container recipients")
		return nil, qs.containerError(err)
	}
	query, values := qs.Query()
	err := db.QueryRow(query, values...).Scan(recipients...)
	if err != nil {
		return nil, qs.dbError(err)
	}
	if qs.conType == containers.Map {
		values := Values{}
		for i, name := range qs.columns {
			values[name] = reflect.ValueOf(recipients[i]).Elem()
		}
		container = values
	}
	instance := &Instance{container, qs.conType, qs.model}
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
