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

func (e PostgresEngine) Stop() error {
	return e.db.Close()
}

func (e PostgresEngine) TxSupport() bool {
	return true
}

func (e PostgresEngine) BeginTx() (Engine, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	e.tx = tx
	return e, nil
}

func (e PostgresEngine) PrepareMigrations() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS gomodels_migration (
		  "id" SERIAL,
		  "app" VARCHAR(50) NOT NULL,
		  "name" VARCHAR(100) NOT NULL,
		  "number" VARCHAR NOT NULL
	)`
	_, err := e.executor().Exec(stmt)
	return err
}
