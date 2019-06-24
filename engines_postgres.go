package gomodels

import (
	"database/sql"
	"fmt"
	"strings"
)

type PostgresEngine struct {
	db *sql.DB
	tx *sql.Tx
}

func (e PostgresEngine) Start(db *Database) (Engine, error) {
	credentials := fmt.Sprintf(
		"dbname=%s user=%s password=%s sslmode=disable",
		db.Name, db.User, db.Password,
	)
	conn, err := sql.Open(db.Driver, credentials)
	if err != nil {
		return nil, err
	}
	e.db = conn
	return e, nil
}

func (e PostgresEngine) Stop() error {
	return e.db.Close()
}

func (e PostgresEngine) DB() *sql.DB {
	return e.db
}

func (e PostgresEngine) Tx() *sql.Tx {
	return e.tx
}

func (e PostgresEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
}

func (e PostgresEngine) CommitTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}
	return e.tx.Commit()
}

func (e PostgresEngine) RollbackTx() error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to roll back")
	}
	return e.tx.Rollback()
}

func (e PostgresEngine) exec(
	stmt string, values ...interface{},
) (sql.Result, error) {
	if e.tx != nil {
		return e.tx.Exec(stmt, values...)
	} else {
		return e.db.Exec(stmt, values...)
	}
}

func (e PostgresEngine) query(
	stmt string, values ...interface{},
) (*sql.Rows, error) {
	if e.tx != nil {
		return e.tx.Query(stmt, values...)
	} else {
		return e.db.Query(stmt, values...)
	}
}

func (e PostgresEngine) queryRow(stmt string, values ...interface{}) *sql.Row {
	if e.tx != nil {
		return e.tx.QueryRow(stmt, values...)
	} else {
		return e.db.QueryRow(stmt, values...)
	}
}

func (e PostgresEngine) CreateTable(tbl string, fields Fields) error {
	columns := make([]string, 0, len(fields))
	for name, field := range fields {
		sqlColumn := fmt.Sprintf(
			"\"%s\" %s", field.DBColumn(name), field.SqlDatatype("postgres"),
		)
		columns = append(columns, sqlColumn)
	}
	stmt := fmt.Sprintf(
		"CREATE TABLE \"%s\" (%s)", tbl, strings.Join(columns, ", "),
	)
	_, err := e.exec(stmt)
	return err
}

func (e PostgresEngine) RenameTable(tbl string, name string) error {
	stmt := fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\"", tbl, name)
	_, err := e.exec(stmt)
	return err
}

func (e PostgresEngine) CopyTable(
	tbl string, name string, cols ...string,
) error {
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

func (e PostgresEngine) DropTable(tbl string) error {
	stmt := fmt.Sprintf("DROP TABLE \"%s\"", tbl)
	_, err := e.exec(stmt)
	return err
}

func (e PostgresEngine) AddIndex(
	tbl string, name string, cols ...string,
) error {
	stmt := fmt.Sprintf(
		"CREATE INDEX \"%s\" ON \"%s\" (%s)",
		name, tbl, strings.Join(cols, ", "),
	)
	_, err := e.exec(stmt)
	return err
}

func (e PostgresEngine) DropIndex(tbl string, name string) error {
	stmt := fmt.Sprintf("DROP INDEX \"%s\"", name)
	_, err := e.exec(stmt)
	return err
}

func (e PostgresEngine) AddColumns(tbl string, fields Fields) error {
	addColumns := make([]string, 0, len(fields))
	for name, field := range fields {
		addColumn := fmt.Sprintf(
			"ADD COLUMN \"%s\" %s",
			field.DBColumn(name), field.SqlDatatype("postgres"),
		)
		addColumns = append(addColumns, addColumn)
	}
	stmt := fmt.Sprintf(
		"ALTER TABLE \"%s\" %s", tbl, strings.Join(addColumns, ", "),
	)
	_, err := e.exec(stmt)
	return err
}

func (e PostgresEngine) DropColumns(tbl string, columns ...string) error {
	dropColumns := make([]string, 0, len(columns))
	for _, name := range columns {
		dropColumns = append(
			dropColumns, fmt.Sprintf("DROP COLUMN \"%s\"", name),
		)
	}
	stmt := fmt.Sprintf(
		"ALTER TABLE %s %s", tbl, strings.Join(dropColumns, ", "),
	)
	_, err := e.exec(stmt)
	return err
}

func (e PostgresEngine) SelectStmt(
	m *Model, c Conditioner, fields ...string,
) (string, []interface{}) {
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
		pred, values := c.Predicate("postgres", 1)
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		return stmt, values
	} else {
		return stmt, nil
	}
}

func (e PostgresEngine) GetRows(
	m *Model, c Conditioner, start int64, end int64, fields ...string,
) (*sql.Rows, error) {
	stmt, values := e.SelectStmt(m, c, fields...)
	if end > 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, end-start)
	} else if start > 0 {
		stmt += " LIMIT ALL"
	}
	if start > 0 {
		stmt = fmt.Sprintf("%s OFFSET %d", stmt, start)
	}
	return e.query(stmt, values...)
}

func (e PostgresEngine) InsertRow(
	model *Model, container Container, fields ...string,
) (int64, error) {
	cols := make([]string, 0, len(model.fields))
	vals := make([]interface{}, 0, len(model.fields))
	placeholders := make([]string, 0, len(model.fields))
	allFields := len(fields) == 0
	index := 1
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
				placeholders = append(placeholders, fmt.Sprintf("$%d", index))
				index += 1
			}
		}
	}
	stmt := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s) RETURNING \"%s\"",
		model.Table(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		model.pk,
	)
	var pk int64
	row := e.queryRow(stmt, vals...)
	err := row.Scan(&pk)
	if err != nil {
		return pk, err
	}
	return pk, nil
}

func (e PostgresEngine) UpdateRows(
	model *Model, cont Container, conditioner Conditioner, fields ...string,
) (int64, error) {
	vals := make([]interface{}, 0, len(model.fields))
	cols := make([]string, 0, len(model.fields))
	allFields := len(fields) == 0
	index := 1
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
				col := fmt.Sprintf(
					"\"%s\" = $%d", field.DBColumn(name), index,
				)
				cols = append(cols, col)
				vals = append(vals, value)
				index += 1
			}
		}
	}
	stmt := fmt.Sprintf(
		"UPDATE \"%s\" SET %s", model.Table(), strings.Join(cols, ", "),
	)
	if conditioner != nil {
		pred, pVals := conditioner.Predicate("postgres", index)
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

func (e PostgresEngine) DeleteRows(model *Model, c Conditioner) (int64, error) {
	var values []interface{}
	stmt := fmt.Sprintf("DELETE FROM %s", model.Table())
	if c != nil {
		pred, vals := c.Predicate("postgres", 1)
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

func (e PostgresEngine) CountRows(model *Model, c Conditioner) (int64, error) {
	var values []interface{}
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s", model.Table())
	if c != nil {
		pred, vals := c.Predicate("postgres", 1)
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

func (e PostgresEngine) Exists(model *Model, c Conditioner) (bool, error) {
	var values []interface{}
	stmt := fmt.Sprintf(
		"SELECT EXISTS (SELECT %s FROM %s)", model.pk, model.Table(),
	)
	if c != nil {
		pred, vals := c.Predicate("postgres", 1)
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