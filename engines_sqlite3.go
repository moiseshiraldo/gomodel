package gomodels

import (
	"database/sql"
	"fmt"
	"strings"
)

type SqliteEngine struct {
	baseSQLEngine
}

func (e SqliteEngine) Start(db Database) (Engine, error) {
	conn, err := sql.Open(db.Driver, db.Name)
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

func (e SqliteEngine) Stop() error {
	return e.db.Close()
}

func (e SqliteEngine) TxSupport() bool {
	return true
}

func (e SqliteEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
}

func (e SqliteEngine) PrepareMigrations() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS gomodels_migration (
		  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		  "app" VARCHAR(50) NOT NULL,
		  "name" VARCHAR(100) NOT NULL,
		  "number" VARCHAR NOT NULL
	)`
	_, err := e.executor().Exec(stmt)
	return err
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

func (e SqliteEngine) DropColumns(
	old *Model,
	new *Model,
	fields ...string,
) error {
	newFields := new.Fields()
	oldFields := old.Fields()
	keepCols := make([]string, 0, len(oldFields)-len(fields))
	for name, field := range newFields {
		keepCols = append(keepCols, field.DBColumn(name))
	}
	name := old.Table() + "__new"
	if err := e.copyTable(old, name, keepCols...); err != nil {
		return err
	}
	if err := e.DropTable(old); err != nil {
		return err
	}
	old.meta.Table = name
	if err := e.RenameTable(old, new); err != nil {
		return err
	}
	for idxName, fields := range new.Indexes() {
		if err := e.AddIndex(new, idxName, fields...); err != nil {
			return err
		}
	}
	return nil
}
