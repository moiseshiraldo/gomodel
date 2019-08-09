package gomodels

import (
	"database/sql"
	"fmt"
)

type MockedEngineResults struct {
	BeginTx           error
	CommitTx          error
	RollbackTx        error
	PrepareMigrations error
	GetMigrations     struct {
		Rows *sql.Rows
		Err  error
	}
	SaveMigration   error
	DeleteMigration error
	CreateTable     error
	RenameTable     error
	CopyTable       error
	DropTable       error
	AddIndex        error
	DropIndex       error
	AddColumns      error
	DropColumns     error
	SelectQuery     struct {
		Query Query
		Err   error
	}
	GetRows struct {
		Rows *sql.Rows
		Err  error
	}
	InsertRow struct {
		Id  int64
		Err error
	}
	UpdateRows struct {
		Number int64
		Err    error
	}
	DeleteRows struct {
		Number int64
		Err    error
	}
	CountRows struct {
		Number int64
		Err    error
	}
	Exists struct {
		Result bool
		Err    error
	}
}

type MockedEngineArgs struct {
	SaveMigration struct {
		App    string
		Number int
		Name   string
	}
	DeleteMigration struct {
		App    string
		Number int
	}
	CreateTable struct {
		Table  string
		Fields Fields
	}
	RenameTable struct {
		Table string
		Name  string
	}
	CopyTable struct {
		Table   string
		Name    string
		Columns []string
	}
	DropTable string
	AddIndex  struct {
		Table   string
		Name    string
		Columns []string
	}
	DropIndex struct {
		Table string
		Name  string
	}
	AddColumns struct {
		Table  string
		Fields Fields
	}
	DropColumns struct {
		Table   string
		Columns []string
	}
	SelectQuery struct {
		Model       *Model
		Conditioner Conditioner
		Fields      []string
	}
	GetRows struct {
		Model       *Model
		Conditioner Conditioner
		Start       int64
		End         int64
		Fields      []string
	}
	InsertRow struct {
		Model     *Model
		Container Container
		Fields    []string
	}
	UpdateRows struct {
		Model       *Model
		Container   Container
		Conditioner Conditioner
		Fields      []string
	}
	DeleteRows struct {
		Model       *Model
		Conditioner Conditioner
	}
	CountRows struct {
		Model       *Model
		Conditioner Conditioner
	}
	Exists struct {
		Model       *Model
		Conditioner Conditioner
	}
}

type MockedEngine struct {
	calls   map[string]int
	Args    *MockedEngineArgs
	Results *MockedEngineResults
	tx      bool
}

func (e MockedEngine) Start(db Database) (Engine, error) {
	e.calls = make(map[string]int)
	e.Results = &MockedEngineResults{}
	return e, nil
}

func (e MockedEngine) Stop() error {
	e.calls["Stop"] += 1
	return nil
}

func (e MockedEngine) DB() *sql.DB {
	e.calls["DB"] += 1
	return nil
}

func (e MockedEngine) Tx() *sql.Tx {
	e.calls["Tx"] += 1
	return nil
}

func (e MockedEngine) BeginTx() (Engine, error) {
	e.calls["BeginTx"] += 1
	e.tx = true
	return e, e.Results.BeginTx
}

func (e MockedEngine) CommitTx() error {
	e.calls["CommitTx"] += 1
	if !e.tx {
		return fmt.Errorf("no transaction to commit")
	}
	return e.Results.CommitTx
}

func (e MockedEngine) RollbackTx() error {
	e.calls["RollbackTx"] += 1
	if !e.tx {
		return fmt.Errorf("no transaction to roll back")
	}
	return e.Results.RollbackTx
}

func (e MockedEngine) PrepareMigrations() error {
	e.calls["PrepareMigrations"] += 1
	return e.Results.PrepareMigrations
}

func (e MockedEngine) GetMigrations() (*sql.Rows, error) {
	e.calls["GetMigrations"] += 1
	return e.Results.GetMigrations.Rows, e.Results.GetMigrations.Err
}

func (e MockedEngine) SaveMigration(app string, number int, name string) error {
	e.calls["SaveMigration"] += 1
	return e.Results.SaveMigration
}

func (e MockedEngine) DeleteMigration(app string, number int) error {
	e.calls["DeleteMigration"] += 1
	return e.Results.DeleteMigration
}

func (e MockedEngine) CreateTable(tbl string, fields Fields) error {
	e.calls["CreateTable"] += 1
	return e.Results.CreateTable
}

func (e MockedEngine) RenameTable(tbl string, name string) error {
	e.calls["RenameTable"] += 1
	return e.Results.RenameTable
}

func (e MockedEngine) CopyTable(tbl string, name string, cols ...string) error {
	e.calls["CopyTable"] += 1
	return e.Results.CopyTable
}

func (e MockedEngine) DropTable(tbl string) error {
	e.calls["DropTable"] += 1
	return e.Results.DropTable
}

func (e MockedEngine) AddIndex(tbl string, name string, cols ...string) error {
	e.calls["AddIndex"] += 1
	return e.Results.AddIndex
}

func (e MockedEngine) DropIndex(tbl string, name string) error {
	e.calls["DropIndex"] += 1
	return e.Results.DropIndex
}

func (e MockedEngine) AddColumns(tbl string, fields Fields) error {
	e.calls["AddColumns"] += 1
	return e.Results.AddColumns
}

func (e MockedEngine) DropColumns(tbl string, columns ...string) error {
	e.calls["DropColumns"] += 1
	return e.Results.DropColumns
}

func (e MockedEngine) SelectQuery(
	m *Model,
	c Conditioner,
	fields ...string,
) (Query, error) {
	e.calls["SelectQuery"] += 1
	return e.Results.SelectQuery.Query, e.Results.SelectQuery.Err
}

func (e MockedEngine) GetRows(
	m *Model,
	c Conditioner,
	start int64,
	end int64,
	fields ...string,
) (*sql.Rows, error) {
	e.calls["GetRows"] += 1
	return e.Results.GetRows.Rows, e.Results.GetRows.Err
}

func (e MockedEngine) InsertRow(
	model *Model,
	container Container,
	fields ...string,
) (int64, error) {
	e.calls["InsertRow"] += 1
	return e.Results.InsertRow.Id, e.Results.InsertRow.Err
}

func (e MockedEngine) UpdateRows(
	model *Model,
	cont Container,
	conditioner Conditioner,
	fields ...string,
) (int64, error) {
	e.calls["UpdateRows"] += 1
	return e.Results.UpdateRows.Number, e.Results.UpdateRows.Err
}

func (e MockedEngine) DeleteRows(model *Model, c Conditioner) (int64, error) {
	e.calls["DeleteRows"] += 1
	return e.Results.DeleteRows.Number, e.Results.DeleteRows.Err
}

func (e MockedEngine) CountRows(model *Model, c Conditioner) (int64, error) {
	e.calls["CountRows"] += 1
	return e.Results.CountRows.Number, e.Results.CountRows.Err
}

func (e MockedEngine) Exists(model *Model, c Conditioner) (bool, error) {
	e.calls["Exists"] += 1
	return e.Results.Exists.Result, e.Results.Exists.Err
}
