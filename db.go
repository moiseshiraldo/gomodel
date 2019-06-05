package gomodels

import (
	"database/sql"
	"fmt"
)

type Database struct {
	Driver string
	Name   string
	conn   *sql.DB
}

func (db Database) Conn() *sql.DB {
	return db.conn
}

var Databases = map[string]Database{}

func Start(options map[string]Database) error {
	for name, db := range options {
		conn, err := sql.Open(db.Driver, db.Name)
		if err != nil {
			fmt.Printf("%+v", err)
			return &DatabaseError{name, ErrorTrace{Err: err}}
		}
		db.conn = conn
		Databases[name] = db
	}
	return nil
}

func Stop() error {
	var err error
	for name, db := range Databases {
		if dbErr := db.conn.Close(); dbErr != nil {
			err = &DatabaseError{name, ErrorTrace{Err: dbErr}}
		}
	}
	return err
}
