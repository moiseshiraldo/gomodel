package gomodels

import (
	"fmt"
	"reflect"
)

type Query struct {
	Stmt string
	Args []interface{}
}

type QuerySet interface {
	Model() *Model
	Container() Container
	SetContainer(c Container) QuerySet
	Filter(c Conditioner) QuerySet
	Exclude(c Conditioner) QuerySet
	Query() (Query, error)
	Load() ([]*Instance, error)
	Get(c Conditioner) (*Instance, error)
	Slice(start int64, end int64) ([]*Instance, error)
	Exists() (bool, error)
	Count() (int64, error)
	Update(values Container) (int64, error)
	Delete() (int64, error)
}

type GenericQuerySet struct {
	model     *Model
	container Container
	database  string
	columns   []string
	cond      Conditioner
}

func (qs GenericQuerySet) trace(err error) ErrorTrace {
	return ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
}

func (qs GenericQuerySet) dbError(err error) error {
	return &DatabaseError{qs.database, qs.trace(err)}
}

func (qs GenericQuerySet) containerError(err error) error {
	return &ContainerError{qs.trace(err)}
}

func (qs GenericQuerySet) addConditioner(c Conditioner) GenericQuerySet {
	if qs.cond == nil {
		if cond, ok := c.(Q); ok {
			qs.cond = condChain{root: cond}
		} else {
			qs.cond = c
		}
	} else {
		qs.cond = qs.cond.And(c)
	}
	return qs
}

func (qs GenericQuerySet) Query() (Query, error) {
	db, ok := databases[qs.database]
	if !ok {
		return Query{}, qs.dbError(fmt.Errorf("db not found"))
	}
	return db.SelectQuery(qs.model, qs.cond, qs.columns...)
}

func (qs GenericQuerySet) Model() *Model {
	return qs.model
}

func (qs GenericQuerySet) Container() Container {
	if b, ok := qs.container.(Builder); ok {
		return b.New()
	} else {
		ct := reflect.TypeOf(qs.container)
		if ct.Kind() == reflect.Ptr {
			ct = ct.Elem()
		}
		return reflect.New(ct).Interface()
	}
}

func (qs GenericQuerySet) SetContainer(container Container) QuerySet {
	if isValidContainer(container) {
		qs.container = container
	} else {
		qs.container = nil
	}
	return qs
}

func (qs GenericQuerySet) Filter(c Conditioner) QuerySet {
	return qs.addConditioner(c)
}

func (qs GenericQuerySet) Exclude(c Conditioner) QuerySet {
	if qs.cond == nil {
		qs.cond = Q{}
	}
	qs.cond = qs.cond.AndNot(c)
	return qs
}

func (qs GenericQuerySet) load(start int64, end int64) ([]*Instance, error) {
	if start < 0 || end != -1 && start >= end || end < -1 {
		err := fmt.Errorf("invalid slice indexes: %d %d", start, end)
		return nil, &QuerySetError{qs.trace(err)}
	}
	result := []*Instance{}
	db, ok := databases[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	if qs.container == nil {
		return nil, qs.containerError(fmt.Errorf("invalid container"))
	}
	container := qs.Container()
	recipients := getRecipients(container, qs.columns, qs.model)
	if len(recipients) != len(qs.columns) {
		err := fmt.Errorf("invalid container recipients")
		return nil, qs.containerError(err)
	}
	rows, err := db.GetRows(qs.model, qs.cond, start, end, qs.columns...)
	if err != nil {
		return nil, qs.dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		container = qs.Container()
		if _, ok := container.(Setter); !ok {
			recipients = getRecipients(container, qs.columns, qs.model)
		}
		err := rows.Scan(recipients...)
		if err != nil {
			return nil, qs.containerError(err)
		}
		instance := &Instance{qs.model, container}
		if _, ok := container.(Setter); ok {
			for i, name := range qs.columns {
				val := reflect.Indirect(
					reflect.ValueOf(recipients[i]),
				).Interface()
				instance.Set(name, val)
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

func (qs GenericQuerySet) Load() ([]*Instance, error) {
	return qs.load(0, -1)
}

func (qs GenericQuerySet) Slice(start int64, end int64) ([]*Instance, error) {
	return qs.load(start, end)
}

func (qs GenericQuerySet) Get(c Conditioner) (*Instance, error) {
	qs = qs.addConditioner(c)
	db, ok := databases[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	container := qs.Container()
	recipients := getRecipients(container, qs.columns, qs.model)
	if len(recipients) != len(qs.columns) {
		err := fmt.Errorf("invalid container recipients")
		return nil, qs.containerError(err)
	}
	rows, err := db.GetRows(qs.model, qs.cond, 0, 2, qs.columns...)
	if err != nil {
		return nil, qs.dbError(err)
	}
	defer rows.Close()
	if err != nil {
		return nil, qs.dbError(err)
	}
	n := 0
	for rows.Next() {
		if n > 0 {
			err := fmt.Errorf("get query returned multiple objects")
			return nil, &MultipleObjectsError{qs.trace(err)}
		}
		err := rows.Scan(recipients...)
		if err != nil {
			return nil, qs.containerError(err)
		}
		n += 1
	}
	if n == 0 {
		err := fmt.Errorf("object does not exist")
		return nil, &ObjectNotFoundError{qs.trace(err)}
	}
	instance := &Instance{qs.model, container}
	if _, ok := container.(Setter); ok {
		for i, name := range qs.columns {
			val := reflect.Indirect(
				reflect.ValueOf(recipients[i]),
			).Interface()
			instance.Set(name, val)
		}
	}
	return instance, nil
}

func (qs GenericQuerySet) Exists() (bool, error) {
	db, ok := databases[qs.database]
	if !ok {
		return false, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	exists, err := db.Exists(qs.model, qs.cond)
	if err != nil {
		return false, qs.dbError(err)
	}
	return exists, nil
}

func (qs GenericQuerySet) Count() (int64, error) {
	db, ok := databases[qs.database]
	if !ok {
		return 0, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	count, err := db.CountRows(qs.model, qs.cond)
	if err != nil {
		return 0, qs.dbError(err)
	}
	return count, nil
}

func (qs GenericQuerySet) Update(values Container) (int64, error) {
	db, ok := databases[qs.database]
	if !ok {
		return 0, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	if !isValidContainer(values) {
		err := fmt.Errorf("invalid values container")
		return 0, qs.containerError(err)
	}
	rows, err := db.UpdateRows(qs.model, values, qs.cond)
	if err != nil {
		return 0, qs.dbError(err)
	}
	return rows, nil
}

func (qs GenericQuerySet) Delete() (int64, error) {
	db, ok := databases[qs.database]
	if !ok {
		return 0, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	rows, err := db.DeleteRows(qs.model, qs.cond)
	if err != nil {
		return 0, qs.dbError(err)
	}
	return rows, nil
}
