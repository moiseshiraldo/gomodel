package gomodel

import (
	"fmt"
	"reflect"
	"time"
)

// A Query holds the details of a database query.
type Query struct {
	Stmt string        // Stmt is the prepared statement.
	Args []interface{} // Args is the list of values for the statement.
}

// The QuerySet interface represents a collection of objects from the database,
// defining the methods to interact with those objects, narrow down the
// collection and retrieve them.
type QuerySet interface {
	// New returns a QuerySet representing all the objects from the given model.
	//
	// The parent will be the type used for all methods returning another
	// QuerySet.
	New(model *Model, parent QuerySet) QuerySet
	// Wrap takes a GenericQuerySet, encloses it and returns a custom type.
	//
	// This method is used to create custom QuerySets by embedding a
	// GenericQuerySet, whose methods will call Wrap and return the custom type.
	Wrap(qs GenericQuerySet) QuerySet
	// Model returns the QuerySet model.
	Model() *Model
	// WithContainer returns a QuerySet with the given Container type as a base.
	WithContainer(container Container) QuerySet
	// WithDB returns a QuerySet where all database operations will be applied
	// on the database identified by the given name.
	WithDB(database string) QuerySet
	// WithTx returns a QuerySet where all database operations will be applied
	// on the given transaction.
	WithTx(tx *Transaction) QuerySet
	// Filter returns a new QuerySet with the given conditioner applied to
	// the collections of objects represented by this QuerySet.
	Filter(c Conditioner) QuerySet
	// Exclude takes the current collection of objects, excludes the ones
	// matching the given conditioner and returns a new QuerySet.
	Exclude(c Conditioner) QuerySet
	// Only returns a QuerySet that will select only the given fields when
	// loaded.
	Only(fields ...string) QuerySet
	// Query returns the SELECT query details for the current QuerySet.
	Query() (Query, error)
	// Load retrieves the collection of objects represented by the QuerySet from
	// the database, and returns a list of instances.
	Load() ([]*Instance, error)
	// Slice retrieves the collection of objects represented by the QuerySet
	// from the start to the end parameters. If end is -1, it will retrieve
	// all objects from the given start.
	Slice(start int64, end int64) ([]*Instance, error)
	// Get returns an instance representing the single object from the current
	// collection matching the given conditioner.
	//
	// If no object is found, *ObjectNotFoundError is returned.
	//
	// If multiple objects match the conditions, *MultipleObjectsError is
	// returned.
	Get(c Conditioner) (*Instance, error)
	// Exists returns true if the collection of objects represented by the
	// QuerySet matches at least one row in the database.
	Exists() (bool, error)
	// Count returns the number of rows matching the collection of objects
	// represented by the QuerySet.
	Count() (int64, error)
	// Update modifies the database rows matching the collection of objects
	// represented by the QuerySet with the given values.
	Update(values Container) (int64, error)
	// Delete removes the database rows matching the collection of objects
	// represented by the QuerySet.
	Delete() (int64, error)
}

// GenericQuerySet implements the QuerySet interface.
type GenericQuerySet struct {
	model     *Model
	container Container
	base      QuerySet
	database  string
	tx        *Transaction
	fields    []string
	cond      Conditioner
}

// New implements the New method of the QuerySet interface.
func (qs GenericQuerySet) New(m *Model, parent QuerySet) QuerySet {
	fields := make([]string, 0, len(m.fields))
	for name := range m.fields {
		fields = append(fields, name)
	}
	qs.model = m
	qs.container = m.meta.Container
	qs.base = parent
	qs.database = "default"
	qs.fields = fields
	return parent.Wrap(qs)
}

// Wrap implements the Wrap method of the QuerySet interface.
func (qs GenericQuerySet) Wrap(parent GenericQuerySet) QuerySet {
	return parent
}

func (qs GenericQuerySet) trace(err error) ErrorTrace {
	return ErrorTrace{App: qs.model.app, Model: qs.model, Err: err}
}

func (qs GenericQuerySet) dbError(err error) error {
	if qs.tx != nil {
		return &DatabaseError{qs.tx.DB.id, qs.trace(err)}
	}
	return &DatabaseError{qs.database, qs.trace(err)}
}

func (qs GenericQuerySet) containerError(err error) error {
	return &ContainerError{qs.trace(err)}
}

func (qs GenericQuerySet) engine() (Engine, error) {
	if qs.tx != nil {
		return qs.tx.Engine, nil
	}
	db, ok := dbRegistry[qs.database]
	if !ok {
		return nil, qs.dbError(fmt.Errorf("db not found: %s", qs.database))
	}
	return db.Engine, nil
}

func (qs GenericQuerySet) addConditioner(c Conditioner) GenericQuerySet {
	if qs.cond == nil {
		qs.cond = c
	} else {
		qs.cond = qs.cond.And(c)
	}
	return qs
}

// Model implements the Model method of the QuerySet interface.
func (qs GenericQuerySet) Model() *Model {
	return qs.model
}

// WithContainer implements the WithContainer method of the QuerySet interface.
func (qs GenericQuerySet) WithContainer(container Container) QuerySet {
	qs.container = container
	return qs.base.Wrap(qs)
}

// WithDB implements the WithDB method of the QuerySet interface.
func (qs GenericQuerySet) WithDB(database string) QuerySet {
	qs.database = database
	return qs.base.Wrap(qs)
}

// WithTx implements the WithTx method of the QuerySet interface.
func (qs GenericQuerySet) WithTx(tx *Transaction) QuerySet {
	qs.tx = tx
	return qs.base.Wrap(qs)
}

// Filter implements the Filter method of the QuerySet interface.
func (qs GenericQuerySet) Filter(c Conditioner) QuerySet {
	return qs.base.Wrap(qs.addConditioner(c))
}

// Exclude implements the Exclude method of the QuerySet interface.
func (qs GenericQuerySet) Exclude(c Conditioner) QuerySet {
	if qs.cond == nil {
		qs.cond = Q{}
	}
	qs.cond = qs.cond.AndNot(c)
	return qs.base.Wrap(qs)
}

// Only implements the Only method of the QuerySet interface.
func (qs GenericQuerySet) Only(fields ...string) QuerySet {
	qs.fields = fields
	pkFound := false
	for _, name := range qs.fields {
		if name == qs.model.pk || name == "pk" {
			pkFound = true
			break
		}
	}
	if !pkFound {
		qs.fields = append(qs.fields, qs.model.pk)
	}
	return qs.base.Wrap(qs)
}

// Query implements the Query method of the QuerySet interface.
func (qs GenericQuerySet) Query() (Query, error) {
	eng, err := qs.engine()
	if err != nil {
		return Query{}, err
	}
	options := QueryOptions{
		Conditioner: qs.cond,
		Fields:      qs.fields,
	}
	return eng.SelectQuery(qs.model, options)
}

func (qs GenericQuerySet) load(start int64, end int64) ([]*Instance, error) {
	if start < 0 || end != -1 && start >= end || end < -1 {
		err := fmt.Errorf("invalid slice indexes: %d %d", start, end)
		return nil, &QuerySetError{qs.trace(err)}
	}
	result := []*Instance{}
	eng, err := qs.engine()
	if err != nil {
		return nil, err
	}
	if !isValidContainer(qs.container) {
		return nil, qs.containerError(fmt.Errorf("invalid container"))
	}
	container := newContainer(qs.container)
	recipients := getRecipients(container, qs.fields, qs.model)
	if len(recipients) != len(qs.fields) {
		err := fmt.Errorf("invalid container recipients")
		return nil, qs.containerError(err)
	}
	options := QueryOptions{
		Conditioner: qs.cond,
		Fields:      qs.fields,
		Start:       start,
		End:         end,
	}
	rows, err := eng.GetRows(qs.model, options)
	if err != nil {
		return nil, qs.dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		container = newContainer(qs.container)
		if _, ok := container.(Setter); !ok {
			recipients = getRecipients(container, qs.fields, qs.model)
		}
		err := rows.Scan(recipients...)
		if err != nil {
			return nil, qs.containerError(err)
		}
		instance := &Instance{qs.model, container}
		if _, ok := container.(Setter); ok {
			for i, name := range qs.fields {
				val := reflect.Indirect(
					reflect.ValueOf(recipients[i]),
				).Interface()
				instance.Set(name, qs.model.fields[name].Value(val))
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

// Load implements the Load method of the QuerySet interface.
func (qs GenericQuerySet) Load() ([]*Instance, error) {
	return qs.load(0, -1)
}

// Slice implemetns the Slice method of the QuerySet interface.
func (qs GenericQuerySet) Slice(start int64, end int64) ([]*Instance, error) {
	return qs.load(start, end)
}

// Get implements the Get method of the QuerySet interface.
func (qs GenericQuerySet) Get(c Conditioner) (*Instance, error) {
	qs = qs.addConditioner(c)
	eng, err := qs.engine()
	if err != nil {
		return nil, err
	}
	if !isValidContainer(qs.container) {
		return nil, qs.containerError(fmt.Errorf("invalid container"))
	}
	container := newContainer(qs.container)
	recipients := getRecipients(container, qs.fields, qs.model)
	if len(recipients) != len(qs.fields) {
		err := fmt.Errorf("invalid container recipients")
		return nil, qs.containerError(err)
	}
	options := QueryOptions{
		Conditioner: qs.cond,
		Fields:      qs.fields,
		Start:       0,
		End:         2,
	}
	rows, err := eng.GetRows(qs.model, options)
	if err != nil {
		return nil, qs.dbError(err)
	}
	defer rows.Close()
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
		for i, name := range qs.fields {
			val := reflect.Indirect(reflect.ValueOf(recipients[i])).Interface()
			instance.Set(name, qs.model.fields[name].Value(val))
		}
	}
	return instance, nil
}

// Exists implements the Exists method of the QuerySet interface.
func (qs GenericQuerySet) Exists() (bool, error) {
	eng, err := qs.engine()
	if err != nil {
		return false, err
	}
	exists, err := eng.Exists(qs.model, QueryOptions{Conditioner: qs.cond})
	if err != nil {
		return false, qs.dbError(err)
	}
	return exists, nil
}

// Count implements the Count method of the QuerySet interface.
func (qs GenericQuerySet) Count() (int64, error) {
	eng, err := qs.engine()
	if err != nil {
		return 0, err
	}
	count, err := eng.CountRows(qs.model, QueryOptions{Conditioner: qs.cond})
	if err != nil {
		return 0, qs.dbError(err)
	}
	return count, nil
}

// Update implements the Update method of the QuerySet interface.
func (qs GenericQuerySet) Update(container Container) (int64, error) {
	eng, err := qs.engine()
	if err != nil {
		return 0, err
	}
	if !isValidContainer(container) {
		err := fmt.Errorf("invalid values container")
		return 0, qs.containerError(err)
	}
	dbValues := Values{}
	for name, field := range qs.model.fields {
		if field.IsAutoNow() {
			dbValues[name] = time.Now()
		} else if val, ok := getContainerField(container, name); ok {
			dbValues[name] = val
		}
	}
	options := QueryOptions{Conditioner: qs.cond}
	rows, err := eng.UpdateRows(qs.model, dbValues, options)
	if err != nil {
		return 0, qs.dbError(err)
	}
	return rows, nil
}

// Delete implements the Delete method of the QuerySet interface.
func (qs GenericQuerySet) Delete() (int64, error) {
	eng, err := qs.engine()
	if err != nil {
		return 0, err
	}
	rows, err := eng.DeleteRows(qs.model, QueryOptions{Conditioner: qs.cond})
	if err != nil {
		return 0, qs.dbError(err)
	}
	return rows, nil
}
