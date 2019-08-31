package gomodels

import (
	"database/sql"
	"fmt"
)

type PostgresEngine struct {
	baseSQLEngine
}

func (e PostgresEngine) Start(db Database) (Engine, error) {
	credentials := fmt.Sprintf(
		"dbname=%s user=%s password=%s sslmode=disable",
		db.Name, db.User, db.Password,
	)
	conn, err := sql.Open(db.Driver, credentials)
	if err != nil {
		return nil, err
	}
	e.baseSQLEngine = baseSQLEngine{
		db:          conn,
		driver:      "postgres",
		escapeChar:  "\"",
		placeholder: "$",
	}
	return e, nil
}

func (e PostgresEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
}
