package gomodels

import (
	"database/sql"
	"fmt"
)

type Database struct {
	Engine
	Driver   string
	Name     string
	User     string
	Password string
	id       string
}

func (db Database) Conn() *sql.DB {
	return db.Engine.DB()
}

func (db Database) BeginTx() (*Transaction, error) {
	engine, err := db.Engine.BeginTx()
	if err != nil {
		return nil, &DatabaseError{db.id, ErrorTrace{Err: err}}
	}
	return &Transaction{engine, db}, nil
}

func (db Database) Id() string {
	return db.id
}

type Transaction struct {
	Engine
	DB Database
}

func (tx Transaction) Conn() *sql.Tx {
	return tx.Engine.Tx()
}

func (tx Transaction) Commit() error {
	return tx.CommitTx()
}

func (tx Transaction) Rollback() error {
	return tx.RollbackTx()
}

var dbRegistry = map[string]Database{}

func Databases() map[string]Database {
	dbs := map[string]Database{}
	for name, db := range dbRegistry {
		dbs[name] = db
	}
	return dbs
}

func Start(options map[string]Database) error {
	for name, db := range options {
		engine, ok := enginesRegistry[db.Driver]
		if !ok {
			err := fmt.Errorf("unsupported driver: %s", db.Driver)
			return &DatabaseError{name, ErrorTrace{Err: err}}
		}
		eng, err := engine.Start(db)
		if err != nil {
			return &DatabaseError{name, ErrorTrace{Err: err}}
		}
		db.Engine = eng
		db.id = name
		db.Password = ""
		dbRegistry[name] = db
	}
	if _, ok := dbRegistry["default"]; !ok {
		err := fmt.Errorf("missing default database")
		return &DatabaseError{"default", ErrorTrace{Err: err}}
	}
	registry["gomodels"] = &Application{
		name:   "gomodels",
		models: map[string]*Model{},
	}
	return nil
}

func Stop() error {
	var err error
	for name, db := range dbRegistry {
		if dbErr := db.Stop(); dbErr != nil {
			err = &DatabaseError{name, ErrorTrace{Err: dbErr}}
		}
	}
	return err
}

type Rows interface {
	Close() error
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}
