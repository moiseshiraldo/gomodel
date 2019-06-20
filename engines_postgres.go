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

func (e PostgresEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
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
	if e.tx != nil {
		return e.tx.Query(stmt, values...)
	} else {
		return e.db.Query(stmt, values...)
	}
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
	var err error
	if e.tx != nil {
		err = e.tx.QueryRow(stmt, vals...).Scan(&pk)
	} else {
		err = e.db.QueryRow(stmt, vals...).Scan(&pk)
	}
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
	var result sql.Result
	var err error
	if e.tx != nil {
		result, err = e.tx.Exec(stmt, vals...)
	} else {
		result, err = e.db.Exec(stmt, vals...)
	}
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
	var result sql.Result
	var err error
	if e.tx != nil {
		result, err = e.tx.Exec(stmt, values...)
	} else {
		result, err = e.db.Exec(stmt, values...)
	}
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
	var rows int64
	var err error
	if e.tx != nil {
		err = e.tx.QueryRow(stmt, values...).Scan(&rows)
	} else {
		err = e.db.QueryRow(stmt, values...).Scan(&rows)
	}
	if err != nil {
		return 0, err
	}
	return rows, nil
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
	var err error
	if e.tx != nil {
		err = e.tx.QueryRow(stmt, values...).Scan(&exists)
	} else {
		err = e.db.QueryRow(stmt, values...).Scan(&exists)
	}
	if err != nil {
		return false, err
	}
	return exists, nil
}
