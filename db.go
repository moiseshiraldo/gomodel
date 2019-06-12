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
	Conn     *sql.DB
}

type DBSettings map[string]Database

var databases = DBSettings{}

func Databases() DBSettings {
	dbs := DBSettings{}
	for name, db := range databases {
		dbs[name] = db
	}
	return dbs
}

func Start(options DBSettings) error {
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
		db.Conn = conn
		db.Password = ""
		databases[name] = db
	}
	if _, ok := databases["default"]; !ok {
		err := fmt.Errorf("missing default database")
		return &DatabaseError{"default", ErrorTrace{Err: err}}
	}
	return nil
}

func Stop() error {
	var err error
	for name, db := range databases {
		if dbErr := db.Conn.Close(); dbErr != nil {
			err = &DatabaseError{name, ErrorTrace{Err: dbErr}}
		}
	}
	return err
}
