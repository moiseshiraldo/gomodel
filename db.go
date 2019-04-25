package gomodels

import (
	"database/sql"
	"fmt"
)

type Database struct {
	Driver string
	Name   string
}

var Databases = map[string]*sql.DB{}

func Start(options map[string]Database) error {
	for name, db := range options {
		conn, err := sql.Open(db.Driver, db.Name)
		if err != nil {
			fmt.Printf("%+v", err)
			return &DatabaseError{name, ErrorTrace{Err: err}}
		}
		Databases[name] = conn
	}
	return nil
}

func Stop() error {
	var err error
	for name, db := range Databases {
		if dbErr := db.Close(); dbErr != nil {
			err = &DatabaseError{name, ErrorTrace{Err: dbErr}}
		}
	}
	return err
}
