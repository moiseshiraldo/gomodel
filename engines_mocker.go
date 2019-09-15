package gomodel

import (
	"database/sql"
	"fmt"
)

// MockedEngineResults holds the results of the Engine interface methods to
// be returned by a MockedEngine.
type MockedEngineResults struct {
	Stop        error
	TxSupport   bool
	BeginTx     error
	CommitTx    error
	RollbackTx  error
	CreateTable error
	RenameTable error
	CopyTable   error
	DropTable   error
	AddIndex    error
	DropIndex   error
	AddColumns  error
	DropColumns error
	SelectQuery struct {
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

// Reset sets all the results back to zero values.
func (r *MockedEngineResults) Reset() {
	*r = MockedEngineResults{}
}

// MockedEngineArgs holds the arguments of the Engine interface methods when
// they've been called from a MockedEngine.
type MockedEngineArgs struct {
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
		Model  *Model
		Fields []string
	}
	SelectQuery struct {
		Model   *Model
		Options QueryOptions
	}
	GetRows struct {
		Model   *Model
		Options QueryOptions
	}
	InsertRow struct {
		Model  *Model
		Values Values
	}
	UpdateRows struct {
		Model   *Model
		Values  Values
		Options QueryOptions
	}
	DeleteRows struct {
		Model   *Model
		Options QueryOptions
	}
	CountRows struct {
		Model   *Model
		Options QueryOptions
	}
	Exists struct {
		Model   *Model
		Options QueryOptions
	}
}

// Reset sets all the arguments back to zero values.
func (a *MockedEngineArgs) Reset() {
	*a = MockedEngineArgs{}
}

// MockedEngine mocks the Engine interface and can be used to write unit tests
// without having to open a database connection.
type MockedEngine struct {
	calls     map[string]int
	Args      *MockedEngineArgs
	Results   *MockedEngineResults
	operators map[string]string
	tx        bool
}

// Calls returns the number of calls made to method.
func (e MockedEngine) Calls(method string) int {
	return e.calls[method]
}

// Reset sets all calls, results and arguments back to zero values.
func (e MockedEngine) Reset() {
	for key := range e.calls {
		delete(e.calls, key)
	}
	e.Args.Reset()
	e.Results.Reset()
}

// Start implements the Start method of the Engine interface.
func (e MockedEngine) Start(db Database) (Engine, error) {
	e.calls = make(map[string]int)
	e.Args = &MockedEngineArgs{}
	e.Results = &MockedEngineResults{}
	e.operators = map[string]string{
		"=":  "=",
		">":  ">",
		">=": ">=",
		"<":  "<",
		"<=": "<=",
	}
	e.calls["Start"] += 1
	return e, nil
}

// Stop mocks the Stop method of the Engine interface.
func (e MockedEngine) Stop() error {
	e.calls["Stop"] += 1
	return e.Results.Stop
}

// TxSupport mocks the TxSupport of the Engine interface.
func (e MockedEngine) TxSupport() bool {
	e.calls["TxSupport"] += 1
	return e.Results.TxSupport
	return true
}

// DB mocks the DB method of the Engine interface. It always returns nil.
func (e MockedEngine) DB() *sql.DB {
	e.calls["DB"] += 1
	return nil
}

// Tx mocks the Tx method of the Engine interface. It always returns nil.
func (e MockedEngine) Tx() *sql.Tx {
	e.calls["Tx"] += 1
	return nil
}

// BeginTx mocks the BeginTx method of the Engine interface. It returns the
// same MockedEngine.
func (e MockedEngine) BeginTx() (Engine, error) {
	e.calls["BeginTx"] += 1
	e.tx = true
	return e, e.Results.BeginTx
}

// CommitTx mocks the CommitTx method of the Engine interface.
func (e MockedEngine) CommitTx() error {
	e.calls["CommitTx"] += 1
	if !e.tx {
		return fmt.Errorf("no transaction to commit")
	}
	return e.Results.CommitTx
}

// RollbackTx mocks the RollbackTx method of the Engine interface.
func (e MockedEngine) RollbackTx() error {
	e.calls["RollbackTx"] += 1
	if !e.tx {
		return fmt.Errorf("no transaction to roll back")
	}
	return e.Results.RollbackTx
}

// CreateTable mocks the CreateTable method of the Engine interface.
func (e MockedEngine) CreateTable(model *Model, force bool) error {
	e.calls["CreateTable"] += 1
	e.Args.CreateTable = model
	return e.Results.CreateTable
}

// RenameTable mocks the RenameTable method of the Engine interface.
func (e MockedEngine) RenameTable(old *Model, new *Model) error {
	e.calls["RenameTable"] += 1
	e.Args.RenameTable.Old = old
	e.Args.RenameTable.New = new
	return e.Results.RenameTable
}

// DropTable mocks the DropTable method of the Engine interface.
func (e MockedEngine) DropTable(model *Model) error {
	e.calls["DropTable"] += 1
	e.Args.DropTable = model
	return e.Results.DropTable
}

// AddIndex mocks the AddIndex method of the Engine interface.
func (e MockedEngine) AddIndex(m *Model, name string, fields ...string) error {
	e.calls["AddIndex"] += 1
	e.Args.AddIndex.Model = m
	e.Args.AddIndex.Name = name
	e.Args.AddIndex.Fields = fields
	return e.Results.AddIndex
}

// DropIndex mocks the DropIndex method of the Engine interface.
func (e MockedEngine) DropIndex(model *Model, name string) error {
	e.calls["DropIndex"] += 1
	e.Args.DropIndex.Model = model
	e.Args.DropIndex.Name = name
	return e.Results.DropIndex
}

// AddColumns mocks the AddColumns method of the Engine interface.
func (e MockedEngine) AddColumns(model *Model, fields Fields) error {
	e.calls["AddColumns"] += 1
	e.Args.AddColumns.Model = model
	e.Args.AddColumns.Fields = fields
	return e.Results.AddColumns
}

// DropColumns mocks the DropColumns method of the Engine interface.
func (e MockedEngine) DropColumns(model *Model, fields ...string) error {
	e.calls["DropColumns"] += 1
	e.Args.DropColumns.Model = model
	e.Args.DropColumns.Fields = fields
	return e.Results.DropColumns
}

// SelectQuery mocks the SelectQuery method of the Engine interface.
func (e MockedEngine) SelectQuery(m *Model, opt QueryOptions) (Query, error) {
	e.calls["SelectQuery"] += 1
	e.Args.SelectQuery.Model = m
	e.Args.SelectQuery.Options = opt
	return e.Results.SelectQuery.Query, e.Results.SelectQuery.Err
}

// GetRows mocks the GetRows method of the Engine interface.
func (e MockedEngine) GetRows(model *Model, opt QueryOptions) (Rows, error) {
	e.calls["GetRows"] += 1
	e.Args.GetRows.Model = model
	e.Args.GetRows.Options = opt
	return e.Results.GetRows.Rows, e.Results.GetRows.Err
}

// InsertRow mocks the InsertRow method of the Engine interface.
func (e MockedEngine) InsertRow(model *Model, values Values) (int64, error) {
	e.calls["InsertRow"] += 1
	e.Args.InsertRow.Model = model
	e.Args.InsertRow.Values = values
	return e.Results.InsertRow.Id, e.Results.InsertRow.Err
}

// UpdateRows mocks the UpdateRows method of the Engine interface.
func (e MockedEngine) UpdateRows(
	model *Model,
	values Values,
	options QueryOptions,
) (int64, error) {
	e.calls["UpdateRows"] += 1
	e.Args.UpdateRows.Model = model
	e.Args.UpdateRows.Values = values
	e.Args.UpdateRows.Options = options
	return e.Results.UpdateRows.Number, e.Results.UpdateRows.Err
}

// DeleteRows mocks the DeleteRows method of the Engine interface.
func (e MockedEngine) DeleteRows(m *Model, opt QueryOptions) (int64, error) {
	e.calls["DeleteRows"] += 1
	e.Args.DeleteRows.Model = m
	e.Args.DeleteRows.Options = opt
	return e.Results.DeleteRows.Number, e.Results.DeleteRows.Err
}

// CountRows mocks the CountRows method of the Engine interface.
func (e MockedEngine) CountRows(m *Model, opt QueryOptions) (int64, error) {
	e.calls["CountRows"] += 1
	e.Args.CountRows.Model = m
	e.Args.CountRows.Options = opt
	return e.Results.CountRows.Number, e.Results.CountRows.Err
}

// Exists mocks the Exists method of the Engine interface.
func (e MockedEngine) Exists(m *Model, opt QueryOptions) (bool, error) {
	e.calls["Exists"] += 1
	e.Args.Exists.Model = m
	e.Args.Exists.Options = opt
	return e.Results.Exists.Result, e.Results.Exists.Err
}
