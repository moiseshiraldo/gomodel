package benchmark

import (
	"fmt"
	_ "github.com/gwenn/gosqlite" // Loads sqlite driver.
	"github.com/moiseshiraldo/gomodel"
)

// User defines a model to be used on benchmarks.
var User = gomodel.New(
	"User",
	gomodel.Fields{
		"firstName":     gomodel.CharField{MaxLength: 50},
		"lastName":      gomodel.CharField{MaxLength: 50},
		"email":         gomodel.CharField{MaxLength: 100},
		"active":        gomodel.BooleanField{Default: true},
		"superuser":     gomodel.BooleanField{DefaultFalse: true},
		"loginAttempts": gomodel.IntegerField{DefaultZero: true},
	},
	gomodel.Options{},
)

type userContainer struct {
	Id            int
	FirstName     string
	LastName      string
	Email         string
	Active        bool
	Superuser     bool
	LoginAttempts int
}

type userBuilder struct {
	Id            int32
	FirstName     string
	LastName      string
	Email         string
	Active        bool
	Superuser     bool
	LoginAttempts int32
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
	case "superuser":
		return u.Superuser, true
	case "loginAttempts":
		return u.LoginAttempts, true
	default:
		return nil, false
	}
}

func (u *userBuilder) Set(
	key string, val gomodel.Value, field gomodel.Field,
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
	case "superuser":
		u.Superuser = val.(bool)
	case "loginAttempts":
		u.LoginAttempts = val.(int32)
	default:
		err = fmt.Errorf("Field not found")
	}
	return err
}

func (u userBuilder) New() gomodel.Builder {
	return &userBuilder{}
}
