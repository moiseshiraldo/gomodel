package gomodels

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type SqliteEngine struct {
	db *sql.DB
	tx *sql.Tx
}

func (e SqliteEngine) Start(db *Database) (Engine, error) {
	conn, err := sql.Open(db.Driver, db.Name)
	if err != nil {
		return nil, err
	}
	e.db = conn
	return e, nil
}

func (e SqliteEngine) Stop() error {
	return e.db.Close()
}

func (e SqliteEngine) DB() *sql.DB {
	return e.db
}

func (e SqliteEngine) Tx() *sql.Tx {
	return e.tx
}

func (e SqliteEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
}

func (e SqliteEngine) CommitTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}
	return e.tx.Commit()
}

func (e SqliteEngine) RollbackTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to roll back")
	}
	return e.tx.Rollback()
}

func (e SqliteEngine) exec(q Query) (sql.Result, error) {
	if e.tx != nil {
		return e.tx.Exec(q.Stmt, q.Args...)
	} else {
		return e.db.Exec(q.Stmt, q.Args...)
	}
}

func (e SqliteEngine) query(q Query) (*sql.Rows, error) {
	if e.tx != nil {
		return e.tx.Query(q.Stmt, q.Args...)
	} else {
		return e.db.Query(q.Stmt, q.Args...)
	}
}

func (e SqliteEngine) queryRow(q Query) *sql.Row {
	if e.tx != nil {
		return e.tx.QueryRow(q.Stmt, q.Args...)
	} else {
		return e.db.QueryRow(q.Stmt, q.Args...)
	}
}

func (e SqliteEngine) PrepareMigrations() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS gomodels_migration (
		  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		  "app" VARCHAR(50) NOT NULL,
		  "name" VARCHAR(100) NOT NULL,
		  "number" VARCHAR NOT NULL
	)`
	_, err := e.exec(Query{Stmt: stmt})
	return err
}

func (e SqliteEngine) GetMigrations() (*sql.Rows, error) {
	return e.query(Query{Stmt: "SELECT app, number FROM gomodels_migration"})
}

func (e SqliteEngine) SaveMigration(app string, number int, name string) error {
	stmt := "INSERT INTO gomodels_migration(app, number, name) VALUES(?, ?, ?)"
	args := []interface{}{app, number, name}
	_, err := e.exec(Query{Stmt: stmt, Args: args})
	return err
}

func (e SqliteEngine) DeleteMigration(app string, number int) error {
	stmt := "DELETE FROM gomodels_migration WHERE app = ? and number = ?"
	args := []interface{}{app, number}
	_, err := e.exec(Query{Stmt: stmt, Args: args})
	return err
}

func (e SqliteEngine) CreateTable(tbl string, fields Fields) error {
	columns := make([]string, 0, len(fields))
	for name, field := range fields {
		sqlColumn := fmt.Sprintf(
			"\"%s\" %s", field.DBColumn(name), field.SqlDatatype("sqlite3"),
		)
		columns = append(columns, sqlColumn)
	}
	stmt := fmt.Sprintf(
		"CREATE TABLE \"%s\" (%s)", tbl, strings.Join(columns, ", "),
	)
	_, err := e.exec(Query{Stmt: stmt})
	return err
}

func (e SqliteEngine) DropTable(tbl string) error {
	stmt := fmt.Sprintf("DROP TABLE \"%s\"", tbl)
	_, err := e.exec(Query{Stmt: stmt})
	return err
}

func (e SqliteEngine) RenameTable(tbl string, name string) error {
	stmt := fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\"", tbl, name)
	_, err := e.exec(Query{Stmt: stmt})
	return err
}

func (e SqliteEngine) CopyTable(tbl string, name string, cols ...string) error {
	columns := make([]string, 0, len(cols))
	for _, col := range cols {
		columns = append(columns, fmt.Sprintf("\"%s\"", col))
	}
	stmt := fmt.Sprintf(
		"CREATE TABLE \"%s\" AS SELECT %s FROM \"%s\"",
		name, strings.Join(columns, ", "), tbl,
	)
	_, err := e.exec(Query{Stmt: stmt})
	return err
}

func (e SqliteEngine) AddIndex(tbl string, name string, cols ...string) error {
	stmt := fmt.Sprintf(
		"CREATE INDEX \"%s\" ON \"%s\" (%s)",
		name, tbl, strings.Join(cols, ", "),
	)
	_, err := e.exec(Query{Stmt: stmt})
	return err
}

func (e SqliteEngine) DropIndex(tbl string, name string) error {
	stmt := fmt.Sprintf("DROP INDEX \"%s\"", name)
	_, err := e.exec(Query{Stmt: stmt})
	return err
}

func (e SqliteEngine) AddColumns(tbl string, fields Fields) error {
	for name, field := range fields {
		stmt := fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" %s",
			tbl, field.DBColumn(name), field.SqlDatatype("sqlite3"),
		)
		if _, err := e.exec(Query{Stmt: stmt}); err != nil {
			return err
		}
	}
	return nil
}

func (e SqliteEngine) DropColumns(tbl string, columns ...string) error {
	return fmt.Errorf("sqlite engine: drop columns not supported")
}

func (e SqliteEngine) predicate(model *Model, cond Conditioner) (Query, error) {
	conditions := make([]string, 0)
	values := make([]interface{}, 0)
	for condition, value := range cond.Predicate() {
		args := strings.Split(condition, " ")
		name := args[0]
		operator := "="
		if len(args) > 1 {
			operator = args[1]
		}
		if _, ok := model.fields[name]; !ok {
			return Query{}, fmt.Errorf("unknown field %s", name)
		}
		column := model.fields[name].DBColumn(name)
		driverVal, err := model.fields[name].DriverValue(value, "sqlite3")
		if err != nil {
			return Query{}, err
		}
		if operator == "=" && driverVal == nil {
			condition = fmt.Sprintf("\"%s\" IS NULL", column)
		} else {
			condition = fmt.Sprintf("\"%s\" %s ?", column, operator)
			values = append(values, driverVal)
		}
		conditions = append(conditions, condition)
	}
	pred := strings.Join(conditions, " AND ")
	next, isOr, isNot := cond.Next()
	if next != nil {
		operator := "AND"
		if isOr {
			operator = "OR"
		}
		if isNot {
			operator += " NOT"
		}
		nextPred, err := e.predicate(model, next)
		if err != nil {
			return Query{}, err
		}
		pred = fmt.Sprintf("(%s) %s (%s)", pred, operator, nextPred.Stmt)
		values = append(values, nextPred.Args...)
	}
	return Query{pred, values}, nil
}

func (e SqliteEngine) SelectQuery(
	m *Model,
	c Conditioner,
	fields ...string,
) (Query, error) {
	query := Query{}
	columns := make([]string, 0, len(m.fields))
	if len(fields) == 0 {
		for name, field := range m.fields {
			columns = append(
				columns, fmt.Sprintf("\"%s\"", field.DBColumn(name)),
			)
		}
	} else {
		if !fieldInList(m.pk, fields) {
			columns = append(
				columns, fmt.Sprintf("\"%s\"", m.fields[m.pk].DBColumn(m.pk)),
			)
		}
		for _, name := range fields {
			col := name
			if field, ok := m.fields[name]; ok {
				col = field.DBColumn(name)
			}
			columns = append(columns, fmt.Sprintf("\"%s\"", col))
		}
	}
	query.Stmt = fmt.Sprintf(
		"SELECT %s FROM %s", strings.Join(columns, ", "), m.Table(),
	)
	if c != nil {
		pred, err := e.predicate(m, c)
		if err != nil {
			return query, err
		}
		query.Stmt = fmt.Sprintf("%s WHERE %s", query.Stmt, pred.Stmt)
		query.Args = pred.Args
	}
	return query, nil
}

func (e SqliteEngine) GetRows(
	m *Model,
	c Conditioner,
	start int64,
	end int64,
	fields ...string,
) (*sql.Rows, error) {
	query, err := e.SelectQuery(m, c, fields...)
	if err != nil {
		return nil, err
	}
	if end > 0 {
		query.Stmt = fmt.Sprintf("%s LIMIT %d", query.Stmt, end-start)
	} else if start > 0 {
		query.Stmt += " LIMIT -1"
	}
	if start > 0 {
		query.Stmt = fmt.Sprintf("%s OFFSET %d", query.Stmt, start)
	}
	return e.query(query)
}

func (e SqliteEngine) InsertRow(
	model *Model,
	container Container,
	fields ...string,
) (int64, error) {
	cols := make([]string, 0, len(model.fields))
	vals := make([]interface{}, 0, len(model.fields))
	placeholders := make([]string, 0, len(model.fields))
	allFields := len(fields) == 0
	for name, field := range model.fields {
		if !field.IsAuto() && (allFields || fieldInList(name, fields)) {
			var value Value
			if getter, ok := container.(Getter); ok {
				if val, ok := getter.Get(name); ok {
					value = val
				}
			} else if val, ok := getStructField(container, name); ok {
				value = val
			} else if field.IsAutoNowAdd() {
				value = time.Now()
			}
			driverVal, err := field.DriverValue(value, "sqlite3")
			if err != nil {
				return 0, err
			}
			if driverVal != nil {
				cols = append(cols, fmt.Sprintf("\"%s\"", field.DBColumn(name)))
				vals = append(vals, driverVal)
				placeholders = append(placeholders, "?")
			}
		}
	}
	stmt := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		model.Table(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	result, err := e.exec(Query{stmt, vals})
	if err != nil {
		return 0, err
	}
	pk, err := result.LastInsertId()
	if err != nil {
		return pk, err
	}
	return pk, nil
}

func (e SqliteEngine) UpdateRows(
	model *Model,
	cont Container,
	conditioner Conditioner,
	fields ...string,
) (int64, error) {
	vals := make([]interface{}, 0, len(model.fields))
	cols := make([]string, 0, len(model.fields))
	allFields := len(fields) == 0
	for name, field := range model.fields {
		if name != model.pk && (allFields || fieldInList(name, fields)) {
			var value Value
			if getter, ok := cont.(Getter); ok {
				if val, ok := getter.Get(name); ok {
					value = val
				}
			} else if val, ok := getStructField(cont, name); ok {
				value = val
			} else if field.IsAutoNow() {
				value = time.Now()
			}
			driverVal, err := field.DriverValue(value, "sqlite3")
			if err != nil {
				return 0, err
			}
			if driverVal != nil {
				cols = append(
					cols, fmt.Sprintf("\"%s\" = ?", field.DBColumn(name)),
				)
				vals = append(vals, driverVal)
			}
		}
	}
	stmt := fmt.Sprintf(
		"UPDATE \"%s\" SET %s", model.Table(), strings.Join(cols, ", "),
	)
	if conditioner != nil {
		pred, err := e.predicate(model, conditioner)
		if err != nil {
			return 0, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred.Stmt)
		vals = append(vals, pred.Args...)
	}
	result, err := e.exec(Query{stmt, vals})
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func (e SqliteEngine) DeleteRows(model *Model, c Conditioner) (int64, error) {
	query := Query{}
	query.Stmt = fmt.Sprintf("DELETE FROM %s", model.Table())
	if c != nil {
		pred, err := e.predicate(model, c)
		if err != nil {
			return 0, err
		}
		query.Stmt = fmt.Sprintf("%s WHERE %s", query.Stmt, pred.Stmt)
		query.Args = pred.Args
	}
	result, err := e.exec(query)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func (e SqliteEngine) CountRows(model *Model, c Conditioner) (int64, error) {
	query := Query{}
	query.Stmt = fmt.Sprintf("SELECT COUNT(*) FROM %s", model.Table())
	if c != nil {
		pred, err := e.predicate(model, c)
		if err != nil {
			return 0, err
		}
		query.Stmt = fmt.Sprintf("%s WHERE %s", query.Stmt, pred.Stmt)
		query.Args = pred.Args
	}
	var count int64
	row := e.queryRow(query)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (e SqliteEngine) Exists(model *Model, c Conditioner) (bool, error) {
	query := Query{}
	query.Stmt = fmt.Sprintf(
		"SELECT EXISTS (SELECT %s FROM %s)", model.pk, model.Table(),
	)
	if c != nil {
		pred, err := e.predicate(model, c)
		if err != nil {
			return false, err
		}
		query.Stmt = fmt.Sprintf("%s WHERE %s", query.Stmt, pred.Stmt)
		query.Args = pred.Args
	}
	var exists bool
	row := e.queryRow(query)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
