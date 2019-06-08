package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

type QuerySet interface {
	Load() ([]*Instance, error)
	Query() (query string, values []interface{})
	Model() *Model
	Columns() []string
	Container() Container
	SetContainer(c Container) QuerySet
	Filter(c Conditioner) QuerySet
	Get(c Conditioner) (*Instance, error)
	Delete() (int64, error)
}

type GenericQuerySet struct {
	model     *Model
	container Container
	conType   string
	database  string
	columns   []string
	cond      Conditioner
}

func (qs GenericQuerySet) dbError(err error) error {
	trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
	return &DatabaseError{qs.database, trace}
}

func (qs GenericQuerySet) containerError(err error) error {
	trace := ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
	return &ContainerError{trace}
}

func (qs GenericQuerySet) addConditioner(c Conditioner) GenericQuerySet {
	if qs.cond == nil {
		if cond, ok := c.(Q); ok {
			qs.cond = Filter{root: cond}
		} else {
			qs.cond = c
		}
	} else {
		qs.cond = qs.cond.And(c)
	}
	return qs
}

func (qs GenericQuerySet) Query() (string, []interface{}) {
	driver := ""
	db, ok := Databases[qs.database]
	if !ok {
		db, ok = Databases["default"]
	}
	if ok {
		driver = db.Driver
	}
	columns := make([]string, 0, len(qs.columns))
	for _, name := range qs.columns {
		col := name
		if field, ok := qs.model.fields[name]; ok {
			col = field.DBColumn(name)
		}
		columns = append(columns, fmt.Sprintf("\"%s\"", col))
	}
	stmt := fmt.Sprintf(
		"SELECT %s FROM %s", strings.Join(columns, ", "), qs.model.Table(),
	)
	if qs.cond != nil {
		pred, values := qs.cond.Predicate(driver, 1)
		stmt += fmt.Sprintf(" WHERE %s", pred)
		return stmt, values
	} else {
		return stmt, make([]interface{}, 0)
	}
}

func (qs GenericQuerySet) Model() *Model {
	return qs.model
}

func (qs GenericQuerySet) Columns() []string {
	return qs.columns
}

func (qs GenericQuerySet) Container() Container {
	switch qs.conType {
	case containers.Map:
		return Values{}
	case containers.Builder:
		return qs.container.(Builder).New()
	default:
		ct := reflect.TypeOf(qs.container)
		if ct.Kind() == reflect.Ptr {
			ct = ct.Elem()
		}
		return reflect.New(ct).Interface()
	}
}

func (qs GenericQuerySet) SetContainer(container Container) QuerySet {
	conType, _ := getContainerType(container)
	qs.conType = conType
	qs.container = container
	return qs
}

func (qs GenericQuerySet) Filter(c Conditioner) QuerySet {
	return qs.addConditioner(c)
}

func (qs GenericQuerySet) Load() ([]*Instance, error) {
	result := []*Instance{}
	db, ok := Databases[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	if qs.conType == "" {
		return nil, qs.containerError(fmt.Errorf("invalid container"))
	}
	container, recipients := getRecipients(qs, qs.conType)
	if len(recipients) != len(qs.columns) {
		err := fmt.Errorf("invalid container recipients")
		return nil, qs.containerError(err)
	}
	stmt, values := qs.Query()
	rows, err := db.conn.Query(stmt, values...)
	if err != nil {
		return nil, qs.dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		container, recipients = getRecipients(qs, qs.conType)
		err := rows.Scan(recipients...)
		if err != nil {
			return nil, qs.containerError(err)
		}
		instance := &Instance{qs.model, container, qs.conType}
		if qs.conType == containers.Map {
			values := instance.container.(Values)
			for i, name := range qs.columns {
				values[name] = reflect.Indirect(
					reflect.ValueOf(recipients[i]),
				).Interface()
			}
		}
		result = append(result, instance)
	}
	err = rows.Err()
	if err != nil {
		return nil, qs.dbError(err)
	}
	return result, nil
}

func (qs GenericQuerySet) Get(c Conditioner) (*Instance, error) {
	qs = qs.addConditioner(c)
	db, ok := Databases[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	container, recipients := getRecipients(qs, qs.conType)
	if len(recipients) != len(qs.columns) {
		err := fmt.Errorf("invalid container recipients")
		return nil, qs.containerError(err)
	}
	stmt, values := qs.Query()
	err := db.conn.QueryRow(stmt, values...).Scan(recipients...)
	if err != nil {
		return nil, qs.dbError(err)
	}
	instance := &Instance{qs.model, container, qs.conType}
	if qs.conType == containers.Map {
		values := instance.container.(Values)
		for i, name := range qs.columns {
			values[name] = reflect.Indirect(
				reflect.ValueOf(recipients[i]),
			).Interface()
		}
	}
	return instance, nil
}

func (qs GenericQuerySet) Delete() (int64, error) {
	db, ok := Databases[qs.database]
	if !ok {
		return 0, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	stmt := fmt.Sprintf("DELETE FROM %s", qs.model.Table())
	values := make([]interface{}, 0)
	if qs.cond != nil {
		pred, vals := qs.cond.Predicate(db.Driver, 1)
		stmt += fmt.Sprintf(" WHERE %s", pred)
		values = append(values, vals)
	}
	result, err := db.conn.Exec(stmt, values...)
	if err != nil {
		return 0, qs.dbError(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, qs.dbError(err)
	}
	return count, nil
}
