package gomodels

type Engine interface {
	Start(*Database) (Engine, error)
	InsertRow(m *Model, c Container, fields ...string) (int64, error)
	UpdateRows(
		m *Model, c Container, cond Conditioner, fields ...string,
	) (int64, error)
}

var engines = map[string]Engine{
	"sqlite3":  SqliteEngine{},
	"postgres": PostgresEngine{},
}
