package gomodels

import (
	"database/sql"
)

type Migrator interface {
	PrepareMigrations() error
	GetMigrations() (*sql.Rows, error)
	SaveMigration(app string, number int, name string) error
	DeleteMigration(app string, number int) error
	CreateTable(table string, fields Fields) error
	RenameTable(table string, name string) error
	CopyTable(table string, name string, columns ...string) error
	DropTable(table string) error
	AddIndex(table string, name string, columns ...string) error
	DropIndex(table string, name string) error
	AddColumns(table string, fields Fields) error
	DropColumns(table string, columns ...string) error
}

type Engine interface {
	Migrator
	Start(*Database) (Engine, error)
	Stop() error
	DB() *sql.DB
	Tx() *sql.Tx
	BeginTx() (Engine, error)
	CommitTx() error
	RollbackTx() error
	SelectQuery(m *Model, cond Conditioner, fields ...string) (Query, error)
	GetRows(m *Model, c Conditioner, start int64, end int64, fields ...string) (*sql.Rows, error)
	InsertRow(m *Model, c Container, fields ...string) (int64, error)
	UpdateRows(m *Model, c Container, cond Conditioner, fields ...string) (int64, error)
	DeleteRows(m *Model, cond Conditioner) (int64, error)
	CountRows(m *Model, cond Conditioner) (int64, error)
	Exists(m *Model, cond Conditioner) (bool, error)
}

var engines = map[string]Engine{
	"sqlite3":  SqliteEngine{},
	"postgres": PostgresEngine{},
}
