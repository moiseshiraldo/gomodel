package benchmarks

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodels"
)

var User = gomodels.New(
	"User",
	gomodels.Fields{
		"firstName":     gomodels.CharField{MaxLength: 50},
		"lastName":      gomodels.CharField{MaxLength: 50},
		"email":         gomodels.CharField{MaxLength: 100},
		"active":        gomodels.BooleanField{Default: true},
		"superuser":     gomodels.BooleanField{DefaultFalse: true},
		"loginAttempts": gomodels.IntegerField{DefaultZero: true},
	},
	gomodels.Options{},
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
	Id            int64
	FirstName     string
	LastName      string
	Email         string
	Active        bool
	Superuser     bool
	LoginAttempts int64
}

func (u userBuilder) Get(key string) (gomodels.Value, bool) {
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
	key string, val gomodels.Value, field gomodels.Field,
) error {
	var err error
	switch key {
	case "id":
		u.Id = val.(int64)
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
		u.LoginAttempts = val.(int64)
	default:
		err = fmt.Errorf("Field not found")
	}
	return err
}

func (u userBuilder) New() gomodels.Builder {
	return &userBuilder{}
}
