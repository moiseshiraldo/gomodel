package gomodels

import (
	"database/sql"
	"fmt"
)

type MockedEngineResults struct {
	Stop              error
	TxSupport         bool
	BeginTx           error
	CommitTx          error
	RollbackTx        error
	PrepareMigrations error
	GetMigrations     struct {
		Rows Rows
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
		Rows Rows
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

func (r *MockedEngineResults) Reset() {
	*r = MockedEngineResults{}
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
	CreateTable *Model
	RenameTable struct {
		Old *Model
		New *Model
	}
	DropTable *Model
	AddIndex  struct {
		Model  *Model
		Name   string
		Fields []string
	}
	DropIndex struct {
		Model *Model
		Name  string
	}
	AddColumns struct {
		Model  *Model
		Fields Fields
	}
	DropColumns struct {
		Old    *Model
		New    *Model
		Fields []string
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
		Model  *Model
		Values Values
	}
	UpdateRows struct {
		Model       *Model
		Values      Values
		Conditioner Conditioner
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

func (a *MockedEngineArgs) Reset() {
	*a = MockedEngineArgs{}
}

type MockedEngine struct {
	calls   map[string]int
	Args    *MockedEngineArgs
	Results *MockedEngineResults
	tx      bool
}

func (e MockedEngine) Calls(method string) int {
	return e.calls[method]
}

func (e MockedEngine) Start(db Database) (Engine, error) {
	e.calls = make(map[string]int)
	e.Args = &MockedEngineArgs{}
	e.Results = &MockedEngineResults{}
	e.calls["Start"] += 1
	return e, nil
}

func (e MockedEngine) Stop() error {
	e.calls["Stop"] += 1
	return e.Results.Stop
}

func (e MockedEngine) Reset() {
	for key := range e.calls {
		delete(e.calls, key)
	}
	e.Args.Reset()
	e.Results.Reset()
}

func (e MockedEngine) TxSupport() bool {
	e.calls["TxSupport"] += 1
	return e.Results.TxSupport
	return true
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

func (e MockedEngine) GetMigrations() (Rows, error) {
	e.calls["GetMigrations"] += 1
	return e.Results.GetMigrations.Rows, e.Results.GetMigrations.Err
}

func (e MockedEngine) SaveMigration(app string, number int, name string) error {
	e.calls["SaveMigration"] += 1
	e.Args.SaveMigration.App = app
	e.Args.SaveMigration.Number = number
	e.Args.SaveMigration.Name = name
	return e.Results.SaveMigration
}

func (e MockedEngine) DeleteMigration(app string, number int) error {
	e.calls["DeleteMigration"] += 1
	e.Args.DeleteMigration.App = app
	e.Args.DeleteMigration.Number = number
	return e.Results.DeleteMigration
}

func (e MockedEngine) CreateTable(model *Model) error {
	e.calls["CreateTable"] += 1
	e.Args.CreateTable = model
	return e.Results.CreateTable
}

func (e MockedEngine) RenameTable(old *Model, new *Model) error {
	e.calls["RenameTable"] += 1
	e.Args.RenameTable.Old = old
	e.Args.RenameTable.New = new
	return e.Results.RenameTable
}

func (e MockedEngine) DropTable(model *Model) error {
	e.calls["DropTable"] += 1
	e.Args.DropTable = model
	return e.Results.DropTable
}

func (e MockedEngine) AddIndex(m *Model, name string, fields ...string) error {
	e.calls["AddIndex"] += 1
	e.Args.AddIndex.Model = m
	e.Args.AddIndex.Name = name
	e.Args.AddIndex.Fields = fields
	return e.Results.AddIndex
}

func (e MockedEngine) DropIndex(model *Model, name string) error {
	e.calls["DropIndex"] += 1
	e.Args.DropIndex.Model = model
	e.Args.DropIndex.Name = name
	return e.Results.DropIndex
}

func (e MockedEngine) AddColumns(model *Model, fields Fields) error {
	e.calls["AddColumns"] += 1
	e.Args.AddColumns.Model = model
	e.Args.AddColumns.Fields = fields
	return e.Results.AddColumns
}

func (e MockedEngine) DropColumns(
	old *Model,
	new *Model,
	fields ...string,
) error {
	e.calls["DropColumns"] += 1
	e.Args.DropColumns.Old = old
	e.Args.DropColumns.New = new
	e.Args.DropColumns.Fields = fields
	return e.Results.DropColumns
}

func (e MockedEngine) SelectQuery(
	model *Model,
	conditioner Conditioner,
	fields ...string,
) (Query, error) {
	e.calls["SelectQuery"] += 1
	e.Args.SelectQuery.Model = model
	e.Args.SelectQuery.Conditioner = conditioner
	e.Args.SelectQuery.Fields = fields
	return e.Results.SelectQuery.Query, e.Results.SelectQuery.Err
}

func (e MockedEngine) GetRows(
	model *Model,
	conditioner Conditioner,
	start int64,
	end int64,
	fields ...string,
) (Rows, error) {
	e.calls["GetRows"] += 1
	e.Args.GetRows.Model = model
	e.Args.GetRows.Conditioner = conditioner
	e.Args.GetRows.Start = start
	e.Args.GetRows.End = end
	e.Args.GetRows.Fields = fields
	return e.Results.GetRows.Rows, e.Results.GetRows.Err
}

func (e MockedEngine) InsertRow(model *Model, values Values) (int64, error) {
	e.calls["InsertRow"] += 1
	e.Args.InsertRow.Model = model
	e.Args.InsertRow.Values = values
	return e.Results.InsertRow.Id, e.Results.InsertRow.Err
}

func (e MockedEngine) UpdateRows(
	model *Model,
	values Values,
	conditioner Conditioner,
) (int64, error) {
	e.calls["UpdateRows"] += 1
	e.Args.UpdateRows.Model = model
	e.Args.UpdateRows.Values = values
	e.Args.UpdateRows.Conditioner = conditioner
	return e.Results.UpdateRows.Number, e.Results.UpdateRows.Err
}

func (e MockedEngine) DeleteRows(model *Model, c Conditioner) (int64, error) {
	e.calls["DeleteRows"] += 1
	e.Args.DeleteRows.Model = model
	e.Args.DeleteRows.Conditioner = c
	return e.Results.DeleteRows.Number, e.Results.DeleteRows.Err
}

func (e MockedEngine) CountRows(model *Model, c Conditioner) (int64, error) {
	e.calls["CountRows"] += 1
	e.Args.CountRows.Model = model
	e.Args.CountRows.Conditioner = c
	return e.Results.CountRows.Number, e.Results.CountRows.Err
}

func (e MockedEngine) Exists(model *Model, c Conditioner) (bool, error) {
	e.calls["Exists"] += 1
	e.Args.Exists.Model = model
	e.Args.Exists.Conditioner = c
	return e.Results.Exists.Result, e.Results.Exists.Err
}
