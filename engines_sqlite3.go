package gomodels

import (
	"database/sql"
	"fmt"
	"strings"
)

type SqliteEngine struct {
	*sql.DB
}

func (e SqliteEngine) Start(db *Database) (Engine, error) {
	conn, err := sql.Open(db.Driver, db.Name)
	if err != nil {
		return nil, err
	}
	e.DB = conn
	db.Conn = conn
	return e, nil
}

func (e SqliteEngine) SelectStmt(
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
		pred, values := c.Predicate("sqlite3", 1)
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		return stmt, values
	} else {
		return stmt, nil
	}
}

func (e SqliteEngine) GetRows(
	m *Model, c Conditioner, start int64, end int64, fields ...string,
) (*sql.Rows, error) {
	stmt, values := e.SelectStmt(m, c, fields...)
	if end > 0 {
		stmt = fmt.Sprintf("%s LIMIT %d", stmt, end-start)
	} else if start > 0 {
		stmt += " LIMIT -1"
	}
	if start > 0 {
		stmt = fmt.Sprintf("%s OFFSET %d", stmt, start)
	}
	rows, err := e.Query(stmt, values...)
	return rows, err
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
	result, err := e.Exec(stmt, vals...)
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
		pred, pVals := conditioner.Predicate("sqlite3", 0)
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		vals = append(vals, pVals...)
	}
	result, err := e.Exec(stmt, vals...)
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
		pred, vals := c.Predicate("sqlite3", 1)
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		values = vals
	}
	result, err := e.Exec(stmt, values...)
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
		pred, vals := c.Predicate("sqlite3", 1)
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		values = vals
	}
	var rows int64
	err := e.QueryRow(stmt, values...).Scan(&rows)
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func (e SqliteEngine) Exists(model *Model, c Conditioner) (bool, error) {
	var values []interface{}
	stmt := fmt.Sprintf(
		"SELECT EXISTS (SELECT %s FROM %s)", model.pk, model.Table(),
	)
	if c != nil {
		pred, vals := c.Predicate("sqlite3", 1)
		stmt = fmt.Sprintf("%s WHERE %s", stmt, pred)
		values = vals
	}
	var exists bool
	err := e.QueryRow(stmt, values...).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
