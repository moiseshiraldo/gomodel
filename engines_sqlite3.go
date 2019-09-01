package gomodels

import (
	"fmt"
	"strings"
)

type SqliteEngine struct {
	baseSQLEngine
}

func (e SqliteEngine) Start(db Database) (Engine, error) {
	conn, err := openDB(db.Driver, db.Name)
	if err != nil {
		return nil, err
	}
	e.baseSQLEngine = baseSQLEngine{
		db:          conn,
		driver:      "sqlite3",
		escapeChar:  "\"",
		placeholder: "?",
	}
	return e, nil
}

func (e SqliteEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
}

func (e SqliteEngine) copyTable(m *Model, name string, cols ...string) error {
	columns := make([]string, 0, len(cols))
	for _, col := range cols {
		columns = append(columns, fmt.Sprintf("\"%s\"", col))
	}
	stmt := fmt.Sprintf(
		"CREATE TABLE %s AS SELECT %s FROM %s",
		e.escape(name), strings.Join(columns, ", "), e.escape(m.Table()),
	)
	_, err := e.executor().Exec(stmt)
	return err
}

func (e SqliteEngine) AddColumns(model *Model, fields Fields) error {
	for name, field := range fields {
		stmt := fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN %s %s %s",
			e.escape(model.Table()),
			e.escape(field.DBColumn(name)),
			field.DataType("sqlite3"),
			e.sqlColumnOptions(field),
		)
		if _, err := e.executor().Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (e SqliteEngine) DropColumns(model *Model, fields ...string) error {
	oldFields := model.Fields()
	keepCols := make([]string, 0, len(oldFields)-len(fields))
	for _, name := range fields {
		delete(oldFields, name)
	}
	for name, field := range oldFields {
		keepCols = append(keepCols, field.DBColumn(name))
	}
	table := model.Table()
	copyTable := table + "__new"
	if err := e.copyTable(model, copyTable, keepCols...); err != nil {
		return err
	}
	if err := e.DropTable(model); err != nil {
		return err
	}
	copyModel := &Model{meta: Options{Table: copyTable}}
	if err := e.RenameTable(copyModel, model); err != nil {
		return err
	}
	for idxName, fields := range model.Indexes() {
		if err := e.AddIndex(model, idxName, fields...); err != nil {
			return err
		}
	}
	return nil
}

func (e SqliteEngine) GetRows(
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
		query.Stmt += " LIMIT -1"
	}
	if start > 0 {
		query.Stmt = fmt.Sprintf("%s OFFSET %d", query.Stmt, start)
	}
	return e.executor().Query(query.Stmt, query.Args...)
}
