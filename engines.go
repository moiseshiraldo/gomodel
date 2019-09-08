package gomodel

import (
	"database/sql"
	"fmt"
	"strings"
)

type Migrator interface {
	CreateTable(model *Model, force bool) error
	RenameTable(old *Model, new *Model) error
	DropTable(model *Model) error
	AddIndex(model *Model, name string, fields ...string) error
	DropIndex(model *Model, name string) error
	AddColumns(model *Model, fields Fields) error
	DropColumns(model *Model, fields ...string) error
}

type Engine interface {
	Migrator
	Start(Database) (Engine, error)
	Stop() error
	TxSupport() bool
	DB() *sql.DB
	Tx() *sql.Tx
	BeginTx() (Engine, error)
	CommitTx() error
	RollbackTx() error
	SelectQuery(m *Model, cond Conditioner, fields ...string) (Query, error)
	GetRows(m *Model, c Conditioner, start int64, end int64, fields ...string) (Rows, error)
	InsertRow(m *Model, vals Values) (int64, error)
	UpdateRows(m *Model, vals Values, cond Conditioner) (int64, error)
	DeleteRows(m *Model, cond Conditioner) (int64, error)
	CountRows(m *Model, cond Conditioner) (int64, error)
	Exists(m *Model, cond Conditioner) (bool, error)
}

var enginesRegistry = map[string]Engine{
	"sqlite3":  SqliteEngine{},
	"postgres": PostgresEngine{},
	"mocker":   MockedEngine{},
}

type Rows interface {
	Close() error
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}

type sqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type sqlDB interface {
	sqlExecutor
	Begin() (*sql.Tx, error)
	Close() error
}

type sqlTx interface {
	sqlExecutor
	Commit() error
	Rollback() error
}

var openDB = func(driver string, credentials string) (*sql.DB, error) {
	return sql.Open(driver, credentials)
}

var scanRow = func(ex sqlExecutor, dest interface{}, query Query) error {
	row := ex.QueryRow(query.Stmt, query.Args...)
	return row.Scan(dest)
}

type baseSQLEngine struct {
	driver      string
	escapeChar  string
	placeholder string
	db          sqlDB
	tx          sqlTx
}

func (e baseSQLEngine) DB() *sql.DB {
	return e.db.(*sql.DB)
}

func (e baseSQLEngine) Tx() *sql.Tx {
	return e.tx.(*sql.Tx)
}

func (e baseSQLEngine) Stop() error {
	return e.db.Close()
}

func (e baseSQLEngine) TxSupport() bool {
	return true
}

func (e baseSQLEngine) CommitTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}
	return e.tx.Commit()
}

func (e baseSQLEngine) RollbackTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to roll back")
	}
	return e.tx.Rollback()
}

func (e baseSQLEngine) executor() sqlExecutor {
	if e.tx != nil {
		return e.tx
	}
	return e.db
}

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

func (e baseSQLEngine) escape(s string) string {
	return fmt.Sprintf("%[1]s%[2]s%[1]s", e.escapeChar, s)
}

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

func (e baseSQLEngine) RenameTable(old *Model, new *Model) error {
	stmt := fmt.Sprintf(
		"ALTER TABLE %s RENAME TO %s",
		e.escape(old.Table()), e.escape(new.Table()),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

func (e baseSQLEngine) DropTable(model *Model) error {
	stmt := fmt.Sprintf("DROP TABLE %s", e.escape(model.Table()))
	_, err := e.executor().Exec(stmt)
	return err
}

func (e baseSQLEngine) AddIndex(
	model *Model,
	name string,
	fields ...string,
) error {
	modelFields := model.Fields()
	columns := make([]string, 0, len(fields))
	for _, fieldName := range fields {
		column := modelFields[fieldName].DBColumn(fieldName)
		columns = append(columns, fmt.Sprintf("\"%s\"", column))
	}
	stmt := fmt.Sprintf(
		"CREATE INDEX %s ON %s (%s)",
		e.escape(name), e.escape(model.Table()), strings.Join(columns, ", "),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

func (e baseSQLEngine) DropIndex(model *Model, name string) error {
	stmt := fmt.Sprintf("DROP INDEX %s", e.escape(name))
	_, err := e.executor().Exec(stmt)
	return err
}

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

func (e baseSQLEngine) predicate(
	model *Model,
	cond Conditioner,
	pIndex int,
) (Query, error) {
	conditions := make([]string, 0)
	values := make([]interface{}, 0)
	operators := map[string]string{
		"=":  "=",
		">":  ">",
		">=": ">=",
		"<":  "<",
		"<=": "<=",
	}
	root, isChain := cond.Root()
	pred := ""
	if isChain {
		rootPred, err := e.predicate(model, root, pIndex)
		if err != nil {
			return Query{}, err
		}
		pred = rootPred.Stmt
		pIndex += len(rootPred.Args)
		values = append(values, rootPred.Args...)
	} else {
		for condition, value := range cond.Conditions() {
			args := strings.Split(condition, " ")
			name := args[0]
			operator := "="
			if len(args) > 1 {
				op, ok := operators[args[1]]
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
				placeholder := e.placeholder
				if placeholder == "$" {
					placeholder = fmt.Sprintf("%s%d", placeholder, pIndex)
				}
				condition = fmt.Sprintf(
					"%s %s %s", e.escape(column), operator, placeholder,
				)
				values = append(values, driverVal)
				pIndex += 1
			}
			conditions = append(conditions, condition)
		}
		pred = strings.Join(conditions, " AND ")
	}
	next, isOr, isNot := cond.Next()
	if next != nil {
		operator := "AND"
		if isOr {
			operator = "OR"
		}
		if isNot {
			operator += " NOT"
		}
		nextPred, err := e.predicate(model, next, pIndex)
		if err != nil {
			return Query{}, err
		}
		pred = fmt.Sprintf("(%s) %s (%s)", pred, operator, nextPred.Stmt)
		values = append(values, nextPred.Args...)
	}
	return Query{pred, values}, nil
}

func (e baseSQLEngine) SelectQuery(
	m *Model,
	c Conditioner,
	fields ...string,
) (Query, error) {
	query := Query{}
	columns := make([]string, 0, len(m.fields))
	for _, name := range fields {
		field, ok := m.fields[name]
		if !ok {
			return query, fmt.Errorf("unknown field: %s", name)
		}
		columns = append(columns, e.escape(field.DBColumn(name)))
	}
	query.Stmt = fmt.Sprintf(
		"SELECT %s FROM %s", strings.Join(columns, ", "), e.escape(m.Table()),
	)
	if c != nil {
		pred, err := e.predicate(m, c, 1)
		if err != nil {
			return query, err
		}
		query.Stmt = fmt.Sprintf("%s WHERE %s", query.Stmt, pred.Stmt)
		query.Args = pred.Args
	}
	return query, nil
}

func (e baseSQLEngine) GetRows(
	m *Model,
	c Conditioner,
	start int64,
	end int64,
	fields ...string,
) (Rows, error) {
	query, err := e.SelectQuery(m, c, fields...)
	if err != nil {
		return nil, err
	}
	if end > 0 {
		query.Stmt = fmt.Sprintf("%s LIMIT %d", query.Stmt, end-start)
	} else if start > 0 {
		query.Stmt += " LIMIT ALL"
	}
	if start > 0 {
		query.Stmt = fmt.Sprintf("%s OFFSET %d", query.Stmt, start)
	}
	return e.executor().Query(query.Stmt, query.Args...)
}

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
			placeholder := e.placeholder
			if placeholder == "$" {
				placeholder = fmt.Sprintf("%s%d", placeholder, index)
			}
			placeholders = append(placeholders, placeholder)
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

func (e baseSQLEngine) UpdateRows(
	model *Model,
	values Values,
	conditioner Conditioner,
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
		placeholder := e.placeholder
		if placeholder == "$" {
			placeholder = fmt.Sprintf("%s%d", placeholder, index)
		}
		col := fmt.Sprintf(
			"%s = %s", e.escape(field.DBColumn(name)), placeholder,
		)
		cols = append(cols, col)
		vals = append(vals, driverVal)
		index += 1
	}
	stmt := fmt.Sprintf(
		"UPDATE %s SET %s", e.escape(model.Table()), strings.Join(cols, ", "),
	)
	if conditioner != nil {
		pred, err := e.predicate(model, conditioner, index)
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

func (e baseSQLEngine) DeleteRows(model *Model, c Conditioner) (int64, error) {
	stmt := fmt.Sprintf("DELETE FROM %s", e.escape(model.Table()))
	args := make([]interface{}, 0)
	if c != nil {
		pred, err := e.predicate(model, c, 1)
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

func (e baseSQLEngine) CountRows(model *Model, c Conditioner) (int64, error) {
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s", e.escape(model.Table()))
	args := make([]interface{}, 0)
	if c != nil {
		pred, err := e.predicate(model, c, 1)
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

func (e baseSQLEngine) Exists(model *Model, c Conditioner) (bool, error) {
	query, err := e.SelectQuery(model, c, model.pk)
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
