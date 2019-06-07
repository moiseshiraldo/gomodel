package gomodels

import (
	"database/sql"
	"fmt"
)

type Database struct {
	Driver   string
	Name     string
	User     string
	Password string
	conn     *sql.DB
}

func (db Database) Conn() *sql.DB {
	return db.conn
}

var Databases = map[string]Database{}

func Start(options map[string]Database) error {
	for name, db := range options {
		credentials := ""
		switch driver := db.Driver; driver {
		case "sqlite3":
			credentials = db.Name
		case "postgres":
			credentials = fmt.Sprintf(
				"dbname=%s user=%s password=%s sslmode=disable",
				db.Name, db.User, db.Password,
			)
		default:
			err := fmt.Errorf("unsupported driver: %s", driver)
			return &DatabaseError{name, ErrorTrace{Err: err}}
		}
		conn, err := sql.Open(db.Driver, credentials)
		if err != nil {
			return &DatabaseError{name, ErrorTrace{Err: err}}
		}
		db.conn = conn
		Databases[name] = db
	}
	if _, ok := Databases["default"]; !ok {
		err := fmt.Errorf("missing default database")
		return &DatabaseError{"default", ErrorTrace{Err: err}}
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
