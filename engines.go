package gomodels

import (
	"database/sql"
)

type Migrator interface {
	PrepareMigrations() error
	GetMigrations() (Rows, error)
	SaveMigration(app string, number int, name string) error
	DeleteMigration(app string, number int) error
	CreateTable(model *Model) error
	RenameTable(old *Model, new *Model) error
	DropTable(model *Model) error
	AddIndex(model *Model, name string, fields ...string) error
	DropIndex(model *Model, name string) error
	AddColumns(model *Model, fields Fields) error
	DropColumns(old *Model, new *Model, fields ...string) error
}

type Engine interface {
	Migrator
	Start(Database) (Engine, error)
	Stop() error
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

var engines = map[string]Engine{
	"sqlite3":  SqliteEngine{},
	"postgres": PostgresEngine{},
	"mocker":   MockedEngine{},
}
