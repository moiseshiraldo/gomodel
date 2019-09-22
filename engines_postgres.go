package gomodel

import (
	"fmt"
)

// postgresOperators holds the supported comparison operators for postgres.
var postgresOperators = map[string]string{
	"=":  "=",
	">":  ">",
	">=": ">=",
	"<":  "<",
	"<=": "<=",
}

// PostgresEngine implements the Engine interface for the postgres driver.
type PostgresEngine struct {
	baseSQLEngine
}

// Start implements the Start method of the Engine interface.
func (e PostgresEngine) Start(db Database) (Engine, error) {
	credentials := fmt.Sprintf(
		"dbname=%s user=%s password=%s sslmode=disable",
		db.Name, db.User, db.Password,
	)
	conn, err := openDB(db.Driver, credentials)
	if err != nil {
		return nil, err
	}
	e.baseSQLEngine = baseSQLEngine{
		db:          conn,
		driver:      "postgres",
		escapeChar:  "\"",
		pHolderChar: "$",
		operators:   postgresOperators,
	}
	return e, nil
}

// BeginTx implements the BeginTx method of the Engine interface.
func (e PostgresEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
}
