package gomodels

import "database/sql"

type Database struct {
	Engine string
	Name   string
	Host   string
}

var Databases map[string]*sql.DB

func Start(options map[string]Database) error {
	for name, db := range options {
		conn, err := sql.Open(db.Engine, db.Name)
		if err != nil {
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
