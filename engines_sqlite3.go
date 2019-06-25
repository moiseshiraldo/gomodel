package gomodels

import (
	"database/sql"
	"fmt"
	"strings"
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

func (e SqliteEngine) exec(
	stmt string, values ...interface{},
) (sql.Result, error) {
	if e.tx != nil {
		return e.tx.Exec(stmt, values...)
	} else {
		return e.db.Exec(stmt, values...)
	}
}

func (e SqliteEngine) query(
	stmt string, values ...interface{},
) (*sql.Rows, error) {
	if e.tx != nil {
		return e.tx.Query(stmt, values...)
	} else {
		return e.db.Query(stmt, values...)
	}
}

func (e SqliteEngine) queryRow(stmt string, values ...interface{}) *sql.Row {
	if e.tx != nil {
		return e.tx.QueryRow(stmt, values...)
	} else {
		return e.db.QueryRow(stmt, values...)
	}
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
	_, err := e.exec(stmt)
	return err
}

func (e SqliteEngine) DropTable(tbl string) error {
	stmt := fmt.Sprintf("DROP TABLE \"%s\"", tbl)
	_, err := e.exec(stmt)
	return err
}

func (e SqliteEngine) RenameTable(tbl string, name string) error {
	stmt := fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\"", tbl, name)
	_, err := e.exec(stmt)
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
	_, err := e.exec(stmt)
	return err
}

func (e SqliteEngine) AddIndex(tbl string, name string, cols ...string) error {
	stmt := fmt.Sprintf(
		"CREATE INDEX \"%s\" ON \"%s\" (%s)",
		name, tbl, strings.Join(cols, ", "),
	)
	_, err := e.exec(stmt)
	return err
}

func (e SqliteEngine) DropIndex(tbl string, name string) error {
	stmt := fmt.Sprintf("DROP INDEX \"%s\"", name)
	_, err := e.exec(stmt)
	return err
}

func (e SqliteEngine) AddColumns(tbl string, fields Fields) error {
	for name, field := range fields {
		stmt := fmt.Sprintf(
			"ALTER TABLE \"%s\" ADD COLUMN \"%s\" %s",
			tbl, field.DBColumn(name), field.SqlDatatype("sqlite3"),
		)
		if _, err := e.exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (e SqliteEngine) DropColumns(tbl string, columns ...string) error {
	return fmt.Errorf("sqlite engine: drop columns not supported")
}

func (e SqliteEngine) predicate(
	model *Model, cond Conditioner,
) (string, []interface{}, error) {
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
			return "", nil, fmt.Errorf("unkown field %s", name)
		}
		column := model.fields[name].DBColumn(name)
		conditions = append(
			conditions, fmt.Sprintf("\"%s\" %s ?", column, operator),
		)
		values = append(values, value)
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
		nextPred, nextValues, err := e.predicate(model, next)
		if err != nil {
			return "", nil, err
		}
		pred = fmt.Sprintf("(%s) %s (%s)", pred, operator, nextPred)
		values = append(values, nextValues...)
	}
	return pred, values, nil
}

func (e SqliteEngine) SelectStmt(
	m *Model, c Conditioner, fields ...string,
) (string, []interface{}, error) {
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
	stmt := fmt.Sprintf(
		"SELECT %s FROM %s", strings.Join(columns, ", "), m.Table(),
	)
	if c != nil {
		pred, values, err := e.predicate(m, c)
		if err != nil {
			return pred, values, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		return stmt, values, nil
	} else {
		return stmt, nil, nil
	}
}

func (e SqliteEngine) GetRows(
	m *Model, c Conditioner, start int64, end int64, fields ...string,
) (*sql.Rows, error) {
	stmt, values, err := e.SelectStmt(m, c, fields...)
	if err != nil {
		return nil, err
	}
	if end > 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, end-start)
	} else if start > 0 {
		stmt += " LIMIT -1"
	}
	if start > 0 {
		stmt = fmt.Sprintf("%s OFFSET %d", stmt, start)
	}
	return e.query(stmt, values...)
}

func (e SqliteEngine) InsertRow(
	model *Model, container Container, fields ...string,
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
			}
			if value != nil {
				cols = append(cols, fmt.Sprintf("\"%s\"", field.DBColumn(name)))
				vals = append(vals, value)
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
	result, err := e.exec(stmt, vals...)
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
	model *Model, cont Container, conditioner Conditioner, fields ...string,
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
			}
			if value != nil {
				cols = append(
					cols, fmt.Sprintf("\"%s\" = ?", field.DBColumn(name)),
				)
				vals = append(vals, value)
			}
		}
	}
	stmt := fmt.Sprintf(
		"UPDATE \"%s\" SET %s", model.Table(), strings.Join(cols, ", "),
	)
	if conditioner != nil {
		pred, pVals, err := e.predicate(model, conditioner)
		if err != nil {
			return 0, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		vals = append(vals, pVals...)
	}
	result, err := e.exec(stmt, vals...)
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
	var values []interface{}
	stmt := fmt.Sprintf("DELETE FROM %s", model.Table())
	if c != nil {
		pred, vals, err := e.predicate(model, c)
		if err != nil {
			return 0, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		values = vals
	}
	result, err := e.exec(stmt, values...)
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
	var values []interface{}
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s", model.Table())
	if c != nil {
		pred, vals, err := e.predicate(model, c)
		if err != nil {
			return 0, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		values = vals
	}
	var count int64
	row := e.queryRow(stmt, values...)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (e SqliteEngine) Exists(model *Model, c Conditioner) (bool, error) {
	var values []interface{}
	stmt := fmt.Sprintf(
		"SELECT EXISTS (SELECT %s FROM %s)", model.pk, model.Table(),
	)
	if c != nil {
		pred, vals, err := e.predicate(model, c)
		if err != nil {
			return false, err
		}
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		values = vals
	}
	var exists bool
	row := e.queryRow(stmt, values...)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
