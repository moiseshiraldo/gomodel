package gomodel

import (
	"fmt"
)

// A Database holds the details required for the ORM to interact with the db.
type Database struct {
	// Engine is the interface providing the database-abstraction API.
	//
	// It's automatically set from the selected Driver.
	Engine
	// Driver is the name of the database/sql driver that will be used to
	// interact with the database.
	Driver string
	// Name is the name of the database used to open a connection.
	Name string
	// User is the user, if required, to open a db connection.
	User string
	// Password is the password, if required, to open a db connection.
	Password string
	// id is the name of the database in the gomodel registry.
	id string
}

// Id returns the database identifier in the gomodel registry.
func (db Database) Id() string {
	return db.id
}

// BeginTx starts and returns a new database Transaction.
func (db Database) BeginTx() (*Transaction, error) {
	engine, err := db.Engine.BeginTx()
	if err != nil {
		return nil, &DatabaseError{db.id, ErrorTrace{Err: err}}
	}
	return &Transaction{engine, db}, nil
}

// Transaction holds a database transaction.
type Transaction struct {
	// Engine is the interface providing the database-abstraction API.
	Engine
	// DB is the Database where the transaction was created.
	DB Database
}

// Commit commits the transaction.
func (tx Transaction) Commit() error {
	return tx.CommitTx()
}

// Rollback rolls back the transction.
func (tx Transaction) Rollback() error {
	return tx.RollbackTx()
}

// dbRegistry is a global map containing all registered databases.
var dbRegistry = map[string]Database{}

// Databases returns a map with all the registered databases.
func Databases() map[string]Database {
	dbs := map[string]Database{}
	for name, db := range dbRegistry {
		dbs[name] = db
	}
	return dbs
}

// Start opens a db connection for each database in the given map and stores
// them in the db registry. It will panic if any of the selected drivers is
// not supported or the connection fails to open.
func Start(options map[string]Database) error {
	for name, db := range options {
		engine, ok := enginesRegistry[db.Driver]
		if !ok {
			msg := fmt.Sprintf(
				"gomodels: %s: unsupported driver: %s", name, db.Driver,
			)
			panic(msg)
		}
		eng, err := engine.Start(db)
		if err != nil {
			panic(fmt.Sprintf("gomodels: %s: %s:", name, err))
		}
		db.Engine = eng
		db.id = name
		db.Password = ""
		dbRegistry[name] = db
	}
	if _, ok := dbRegistry["default"]; !ok {
		panic("gomodels: missing default database")
	}
	registry["gomodel"] = &Application{
		name:   "gomodel",
		models: map[string]*Model{},
	}
	return nil
}

// Stop close all the db connections and removes them from the db registry.
func Stop() error {
	var err error
	for name, db := range dbRegistry {
		if dbErr := db.Stop(); dbErr != nil {
			err = &DatabaseError{name, ErrorTrace{Err: dbErr}}
		}
	}
	return err
}
