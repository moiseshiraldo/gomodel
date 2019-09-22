package gomodel

import (
	"database/sql"
	"fmt"
	"strings"
)

// QueryOptions holds common arguments for some Engine interface methods.
type QueryOptions struct {
	// Conditioner holds the conditions to be applied on the query.
	Conditioner Conditioner
	// Fields are the columns that will be selected on the query.
	Fields []string
	// Start represents the first row index to be selected, starting at 0.
	Start int64
	// End represents the row index to stop the selection.
	End int64
}

// Engine is the interface providing the database-abstraction API methods.
type Engine interface {
	// Start opens a connection using the given Database details and returns
	// a new Engine holding the connection.
	Start(Database) (Engine, error)
	// Stop closes the db connection held by the engine.
	Stop() error
	// TxSupport indicates whether the engine supports transactions or not.
	TxSupport() bool
	// DB returns the underlying *sql.DB or nil if not applicable.
	DB() *sql.DB
	// BeginTx starts a new transaction and returns a new Engine holding it.
	BeginTx() (Engine, error)
	// Tx returns the underlying *sql.Tx or nil if not applicable.
	Tx() *sql.Tx
	// CommitTx commits the transaction held by the engine.
	CommitTx() error
	// RollbackTx rolls back the transaction held by the engine.
	RollbackTx() error
	// CreateTable creates the table for the given model. If force is true, the
	// method should return an error if the table already exists.
	CreateTable(model *Model, force bool) error
	// RenameTable renames the table from the given old model to the new one.
	RenameTable(old *Model, new *Model) error
	// DropTable drops the table for the given model.
	DropTable(model *Model) error
	// AddIndex creates a new index for the given model and fields.
	AddIndex(model *Model, name string, fields ...string) error
	// DropIndex drops the named index for the given model.
	DropIndex(model *Model, name string) error
	// AddColumns creates the columns for the given fields on the model table.
	AddColumns(model *Model, fields Fields) error
	// DropColumns drops the columns given by fields from the model table.
	DropColumns(model *Model, fields ...string) error
	// SelectQuery returns the SELECT SQL query details for the given model and
	// query options.
	SelectQuery(model *Model, options QueryOptions) (Query, error)
	// GetRows returns the Rows resulting from querying the database with
	// the query options, where options.Start is the first row index (starting
	// at 0) and options.End the last row index (-1 for all rows).
	GetRows(model *Model, options QueryOptions) (Rows, error)
	// InsertRow inserts the given values in the model table.
	InsertRow(model *Model, values Values) (int64, error)
	// UpdateRows updates the model rows selected by the given conditioner with
	// the given values.
	UpdateRows(model *Model, values Values, options QueryOptions) (int64, error)
	// DeleteRows deletes the model rows selected by the given conditioner.
	DeleteRows(model *Model, options QueryOptions) (int64, error)
	// CountRows counts the model rows selected by the given conditioner.
	CountRows(model *Model, options QueryOptions) (int64, error)
	// Exists returns whether any model row exists for the given conditioner.
	Exists(model *Model, options QueryOptions) (bool, error)
}

// enginesRegistry is a global variable mapping the supported drivers to the
// corresopnding engine.
var enginesRegistry = map[string]Engine{
	"sqlite3":  SqliteEngine{},
	"postgres": PostgresEngine{},
	"mocker":   MockedEngine{},
}

// Rows is an interface wrapping the sql.Rows methods, allowing custom types
// to be returned by the Engine GetRows method.
type Rows interface {
	// Close closes the rows.
	Close() error
	// Err returns the error, if any, that was encountered during iteration.
	Err() error
	// Next prepares the next result row for reading with the Scan method. It
	// returns true on success, or false if there is no next result row or an
	// error happened while preparing it.
	Next() bool
	// Scan copies the columns in the current row into the values pointed at
	// by dest. The number of values in dest must be the same as the number of
	// columns in Rows.
	Scan(dest ...interface{}) error
}

// sqlExecutor is an interface wrapping the sql package query execution methods.
type sqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// sqlDB is an interface wrapping the sql.DB methods.
type sqlDB interface {
	sqlExecutor
	Begin() (*sql.Tx, error)
	Close() error
}

// sqlTx is an interface wrapping the sql.Tx methods.
type sqlTx interface {
	sqlExecutor
	Commit() error
	Rollback() error
}

// openDB holds the function to open sql.DB connections.
var openDB = func(driver string, credentials string) (*sql.DB, error) {
	return sql.Open(driver, credentials)
}

// scanRow holds the function to scan single row results.
var scanRow = func(ex sqlExecutor, dest interface{}, query Query) error {
	row := ex.QueryRow(query.Stmt, query.Args...)
	return row.Scan(dest)
}

// baseSQLEngine implements common Engine methods for SQL drivers.
type baseSQLEngine struct {
	driver      string            // Driver name.
	escapeChar  string            // Escape char for columns and table names.
	pHolderChar string            // Placeholder character for values.
	operators   map[string]string // Available comparison operators.
	db          sqlDB             // *sql.DB
	tx          sqlTx             // *sql.Tx
}

// DB implements the DB method of the Engine interface.
func (e baseSQLEngine) DB() *sql.DB {
	return e.db.(*sql.DB)
}

// Tx implements the Tx method of the Engine interface.
func (e baseSQLEngine) Tx() *sql.Tx {
	return e.tx.(*sql.Tx)
}

// Stop implements the Stop method of the Engine interface.
func (e baseSQLEngine) Stop() error {
	return e.db.Close()
}

// TxSupport implements the TxSupport method of the Engine interface.
func (e baseSQLEngine) TxSupport() bool {
	return true
}

// CommitTx implements the CommitTx method of the Engine interface.
func (e baseSQLEngine) CommitTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}
	return e.tx.Commit()
}

// RollbackTx implements the RollbackTx method of the Engine interface.
func (e baseSQLEngine) RollbackTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to roll back")
	}
	return e.tx.Rollback()
}

// executor returns the underlying *sql.DB or *sql.Tx for transactions.
func (e baseSQLEngine) executor() sqlExecutor {
	if e.tx != nil {
		return e.tx
	}
	return e.db
}

// sqlColumnOptions returns the options for a new column.
func (e baseSQLEngine) sqlColumnOptions(field Field) string {
	options := []string{""}
	if !field.IsAuto() {
		if field.IsNull() {
			options = append(options, "NULL")
		} else {
			options = append(options, "NOT NULL")
		}
	}
	if field.IsPK() {
		options = append(options, "PRIMARY KEY")
	} else if field.IsUnique() {
		options = append(options, "UNIQUE")
	}
	return strings.Join(options, " ")
}

// escape returns the escaped given string.
func (e baseSQLEngine) escape(s string) string {
	return fmt.Sprintf("%[1]s%[2]s%[1]s", e.escapeChar, s)
}

// placeholder returns the placeholder string for the given index.
func (e baseSQLEngine) placeholder(index int) string {
	placeholder := e.pHolderChar
	if placeholder == "$" {
		placeholder = fmt.Sprintf("%s%d", placeholder, index)
	}
	return placeholder
}

// CreateTable implements the CreateTable method of the Engine interface.
func (e baseSQLEngine) CreateTable(model *Model, force bool) error {
	fields := model.Fields()
	columns := make([]string, 0, len(fields))
	for name, field := range fields {
		sqlColumn := fmt.Sprintf(
			"%s %s%s",
			e.escape(field.DBColumn(name)),
			field.DataType(e.driver),
			e.sqlColumnOptions(field),
		)
		columns = append(columns, sqlColumn)
	}
	skip := ""
	if !force {
		skip = "IF NOT EXISTS "
	}
	stmt := fmt.Sprintf(
		"CREATE TABLE %s%s (%s)",
		skip, e.escape(model.Table()), strings.Join(columns, ", "),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

// RenameTable implements the RenameTable method of the Engine interface.
func (e baseSQLEngine) RenameTable(old *Model, new *Model) error {
	stmt := fmt.Sprintf(
		"ALTER TABLE %s RENAME TO %s",
		e.escape(old.Table()), e.escape(new.Table()),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

// DropTable implements the DropTable method of the Engine interface.
func (e baseSQLEngine) DropTable(model *Model) error {
	stmt := fmt.Sprintf("DROP TABLE %s", e.escape(model.Table()))
	_, err := e.executor().Exec(stmt)
	return err
}

// AddIndex implements the AddIndex method of the Engine interface.
func (e baseSQLEngine) AddIndex(m *Model, name string, fields ...string) error {
	modelFields := m.Fields()
	columns := make([]string, 0, len(fields))
	for _, fieldName := range fields {
		column := modelFields[fieldName].DBColumn(fieldName)
		columns = append(columns, fmt.Sprintf("\"%s\"", column))
	}
	stmt := fmt.Sprintf(
		"CREATE INDEX %s ON %s (%s)",
		e.escape(name), e.escape(m.Table()), strings.Join(columns, ", "),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

// DropIndex implements the DropIndex method of the Engine interface.
func (e baseSQLEngine) DropIndex(model *Model, name string) error {
	stmt := fmt.Sprintf("DROP INDEX %s", e.escape(name))
	_, err := e.executor().Exec(stmt)
	return err
}

// AddColumns implements the AddColumns method of the Engine interface.
func (e baseSQLEngine) AddColumns(model *Model, fields Fields) error {
	addColumns := make([]string, 0, len(fields))
	for name, field := range fields {

		addColumn := fmt.Sprintf(
			"ADD COLUMN %s %s %s",
			e.escape(field.DBColumn(name)),
			field.DataType(e.driver),
			e.sqlColumnOptions(field),
		)
		addColumns = append(addColumns, addColumn)
	}
	stmt := fmt.Sprintf(
		"ALTER TABLE %s %s",
		e.escape(model.Table()), strings.Join(addColumns, ", "),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

// DropColumns implements the DropColumns method of the Engine interface.
func (e baseSQLEngine) DropColumns(model *Model, fields ...string) error {
	oldFields := model.Fields()
	dropColumns := make([]string, 0, len(fields))
	for _, name := range fields {
		dropColumns = append(
			dropColumns,
			fmt.Sprintf(
				"DROP COLUMN %s", e.escape(oldFields[name].DBColumn(name)),
			),
		)
	}
	stmt := fmt.Sprintf(
		"ALTER TABLE %s %s",
		e.escape(model.Table()), strings.Join(dropColumns, ", "),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

// predicate returns the SQL predicate for the given conditioner.
//
// pIndex is the next index if the value placeholder requires indexing.
func (e baseSQLEngine) predicate(
	model *Model,
	options QueryOptions,
	pIndex int,
) (Query, error) {
	conditions := make([]string, 0)
	values := make([]interface{}, 0)
	root, isChain := options.Conditioner.Root()
	pred := ""
	if isChain {
		rootPred, err := e.predicate(
			model, QueryOptions{Conditioner: root}, pIndex,
		)
		if err != nil {
			return Query{}, err
		}
		pred = rootPred.Stmt
		pIndex += len(rootPred.Args)
		values = append(values, rootPred.Args...)
	} else {
		for condition, value := range options.Conditioner.Conditions() {
			args := strings.Split(condition, " ")
			name := args[0]
			if name == "pk" {
				name = model.pk
			}
			operator := "="
			if len(args) > 1 {
				op, ok := e.operators[args[1]]
				if !ok {
					return Query{}, fmt.Errorf("invalid operator: %s", args[1])
				}
				operator = op
			}
			if _, ok := model.fields[name]; !ok {
				return Query{}, fmt.Errorf("unknown field %s", name)
			}
			column := model.fields[name].DBColumn(name)
			driverVal, err := model.fields[name].DriverValue(value, e.driver)
			if err != nil {
				return Query{}, err
			}
			if operator == "=" && driverVal == nil {
				condition = fmt.Sprintf("%s IS NULL", e.escape(column))
			} else {
				condition = fmt.Sprintf(
					"%s %s %s",
					e.escape(column), operator, e.placeholder(pIndex),
				)
				values = append(values, driverVal)
				pIndex += 1
			}
			conditions = append(conditions, condition)
		}
		pred = strings.Join(conditions, " AND ")
	}
	next, isOr, isNot := options.Conditioner.Next()
	if next != nil {
		nextPred, err := e.predicate(
			model, QueryOptions{Conditioner: next}, pIndex,
		)
		if pred == "" {
			pred = nextPred.Stmt
			if isNot {
				pred = fmt.Sprintf("NOT (%s)", pred)
			}
		} else {
			operator := "AND"
			if isOr {
				operator = "OR"
			}
			if isNot {
				operator += " NOT"
			}
			if err != nil {
				return Query{}, err
			}
			pred = fmt.Sprintf("(%s) %s (%s)", pred, operator, nextPred.Stmt)
		}
		values = append(values, nextPred.Args...)
	}
	return Query{pred, values}, nil
}

// SelectQuery implements the SelectQuery method of the Engine interface.
func (e baseSQLEngine) SelectQuery(m *Model, opt QueryOptions) (Query, error) {
	query := Query{}
	columns := make([]string, 0, len(opt.Fields))
	for _, name := range opt.Fields {
		field, ok := m.fields[name]
		if !ok {
			return query, fmt.Errorf("unknown field: %s", name)
		}
		columns = append(columns, e.escape(field.DBColumn(name)))
	}
	query.Stmt = fmt.Sprintf(
		"SELECT %s FROM %s", strings.Join(columns, ", "), e.escape(m.Table()),
	)
	if opt.Conditioner != nil {
		pred, err := e.predicate(m, opt, 1)
		if err != nil {
			return query, err
		}
		query.Stmt = fmt.Sprintf("%s WHERE %s", query.Stmt, pred.Stmt)
		query.Args = pred.Args
	}
	return query, nil
}

// GetRows implements the GetRows method of the Engine interface.
func (e baseSQLEngine) GetRows(model *Model, opt QueryOptions) (Rows, error) {
	query, err := e.SelectQuery(model, opt)
	if err != nil {
		return nil, err
	}
	if opt.End > 0 {
		query.Stmt = fmt.Sprintf("%s LIMIT %d", query.Stmt, opt.End-opt.Start)
	} else if opt.Start > 0 {
		query.Stmt += " LIMIT ALL"
	}
	if opt.Start > 0 {
		query.Stmt = fmt.Sprintf("%s OFFSET %d", query.Stmt, opt.Start)
	}
	return e.executor().Query(query.Stmt, query.Args...)
}

// InsertRow implements the InsertRow method of the Engine interface.
func (e baseSQLEngine) InsertRow(model *Model, values Values) (int64, error) {
	cols := make([]string, 0, len(model.fields))
	vals := make([]interface{}, 0, len(model.fields))
	placeholders := make([]string, 0, len(model.fields))
	fields := model.Fields()
	index := 1
	for name, val := range values {
		field, ok := fields[name]
		if !ok {
			return 0, fmt.Errorf("unknown field %s", name)
		}
		driverVal, err := field.DriverValue(val, e.driver)
		if err != nil {
			return 0, err
		}
		if driverVal != nil {
			cols = append(cols, e.escape(field.DBColumn(name)))
			vals = append(vals, driverVal)
			placeholders = append(placeholders, e.placeholder(index))
			index += 1
		}
	}
	stmt := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		e.escape(model.Table()),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	if e.driver == "postgres" {
		stmt = fmt.Sprintf(
			"%s RETURNING %s",
			stmt, e.escape(model.fields[model.pk].DBColumn(model.pk)),
		)
		var pk int64
		if err := scanRow(e.executor(), &pk, Query{stmt, vals}); err != nil {
			return pk, err
		}
		return pk, nil
	}
	result, err := e.executor().Exec(stmt, vals...)
	if err != nil {
		return 0, err
	}
	pk, err := result.LastInsertId()
	if err != nil {
		return pk, err
	}
	return pk, nil
}

// UpdateRows implements the UpdateRows method of the Engine interface.
func (e baseSQLEngine) UpdateRows(
	model *Model,
	values Values,
	options QueryOptions,
) (int64, error) {
	vals := make([]interface{}, 0, len(model.fields))
	cols := make([]string, 0, len(model.fields))
	fields := model.Fields()
	index := 1
	for name, val := range values {
		field, ok := fields[name]
		if !ok {
			return 0, fmt.Errorf("unknown field %s", name)
		}
		driverVal, err := field.DriverValue(val, e.driver)
		if err != nil {
			return 0, err
		}
		col := fmt.Sprintf(
			"%s = %s", e.escape(field.DBColumn(name)), e.placeholder(index),
		)
		cols = append(cols, col)
		vals = append(vals, driverVal)
		index += 1
	}
	stmt := fmt.Sprintf(
		"UPDATE %s SET %s", e.escape(model.Table()), strings.Join(cols, ", "),
	)
	if options.Conditioner != nil {
		pred, err := e.predicate(model, options, index)
		if err != nil {
			return 0, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred.Stmt)
		vals = append(vals, pred.Args...)
	}
	result, err := e.executor().Exec(stmt, vals...)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

// DeleteRows implements the DeleteRows method of the engine interface.
func (e baseSQLEngine) DeleteRows(m *Model, opt QueryOptions) (int64, error) {
	stmt := fmt.Sprintf("DELETE FROM %s", e.escape(m.Table()))
	args := make([]interface{}, 0)
	if opt.Conditioner != nil {
		pred, err := e.predicate(m, opt, 1)
		if err != nil {
			return 0, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred.Stmt)
		args = pred.Args
	}
	result, err := e.executor().Exec(stmt, args...)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

// CountRows implement the CountRows method of the Engine interface.
func (e baseSQLEngine) CountRows(m *Model, opt QueryOptions) (int64, error) {
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s", e.escape(m.Table()))
	args := make([]interface{}, 0)
	if opt.Conditioner != nil {
		pred, err := e.predicate(m, opt, 1)
		if err != nil {
			return 0, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred.Stmt)
		args = pred.Args
	}
	var count int64
	if err := scanRow(e.executor(), &count, Query{stmt, args}); err != nil {
		return 0, err
	}
	return count, nil
}

// Exsists implements the Exists method of the Engine interface.
func (e baseSQLEngine) Exists(m *Model, opt QueryOptions) (bool, error) {
	opt.Fields = []string{m.pk}
	query, err := e.SelectQuery(m, opt)
	if err != nil {
		return false, err
	}
	query.Stmt = fmt.Sprintf("SELECT EXISTS (%s)", query.Stmt)
	var exists bool
	if err = scanRow(e.executor(), &exists, query); err != nil {
		return false, err
	}
	return exists, nil
}
