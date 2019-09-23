package benchmark

import (
	"fmt"
	_ "github.com/gwenn/gosqlite" // Loads sqlite driver.
	"github.com/moiseshiraldo/gomodel"
	"time"
)

// User defines a model to be used on benchmarks.
var User = gomodel.New(
	"User",
	gomodel.Fields{
		"firstName":     gomodel.CharField{MaxLength: 50},
		"lastName":      gomodel.CharField{MaxLength: 50},
		"email":         gomodel.CharField{MaxLength: 100},
		"active":        gomodel.BooleanField{Default: true},
		"loginAttempts": gomodel.IntegerField{DefaultZero: true},
		"created":       gomodel.DateTimeField{AutoNowAdd: true},
	},
	gomodel.Options{},
)

type userContainer struct {
	Id            int
	FirstName     string
	LastName      string
	Email         string
	Active        bool
	LoginAttempts int
	Created       string
}

type userBuilder struct {
	Id            int32
	FirstName     string
	LastName      string
	Email         string
	Active        bool
	LoginAttempts int32
	Created       time.Time
}

func (u userBuilder) Get(key string) (gomodel.Value, bool) {
	switch key {
	case "id":
		return u.Id, true
	case "firstName":
		return u.FirstName, true
	case "lastName":
		return u.LastName, true
	case "email":
		return u.Email, true
	case "active":
		return u.Active, true
	case "loginAttempts":
		return u.LoginAttempts, true
	case "created":
		return u.Created, true
	default:
		return nil, false
	}
}

func (u *userBuilder) Set(
	key string,
	val gomodel.Value,
	field gomodel.Field,
) error {
	var err error
	switch key {
	case "id":
		u.Id = val.(int32)
	case "firstName":
		u.FirstName = val.(string)
	case "lastName":
		u.LastName = val.(string)
	case "email":
		u.Email = val.(string)
	case "active":
		u.Active = val.(bool)
	case "loginAttempts":
		u.LoginAttempts = val.(int32)
	case "created":
		u.Created = val.(time.Time)
	default:
		err = fmt.Errorf("Field not found")
	}
	return err
}

func (u userBuilder) New() gomodel.Builder {
	return &userBuilder{}
}
