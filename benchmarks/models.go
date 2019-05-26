package benchmarks

import (
    "github.com/moiseshiraldo/gomodels"
    "fmt"
    _ "github.com/gwenn/gosqlite"
)

var User = gomodels.New(
    "User",
    gomodels.Fields{
        "firstName": gomodels.CharField{MaxLength: 50},
        "lastName": gomodels.CharField{MaxLength: 50},
        "email": gomodels.CharField{MaxLength: 100},
        "active": gomodels.BooleanField{Default: true},
        "superuser": gomodels.BooleanField{DefaultFalse: true},
        "loginAttempts": gomodels.IntegerField{DefaultZero: true},
    },
    gomodels.Options{},
)

type userContainer struct {
    Id int
    FirstName string
    LastName string
    Email string
    Active bool
    Superuser bool
    LoginAttempts int
}

type userBuilder struct {
    Id int
    FirstName string
    LastName string
    Email string
    Active bool
    Superuser bool
    LoginAttempts int
}

func (u userBuilder) Get(field string) (gomodels.Value, bool) {
    switch field {
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

func (u *userBuilder) Set(field string, val gomodels.Value) error {
    var err error
    switch field {
	case "id":
		u.Id = val.(int)
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
        u.LoginAttempts = val.(int)
	default:
		err = fmt.Errorf("Field not found")
	}
    return err
}

func (u userBuilder) New() gomodels.Builder {
	return &userBuilder{}
}

func (u *userBuilder) Recipients(columns []string) []interface{} {
    recipients := make([]interface{}, 0, len(columns))
	for _, col := range columns {
        switch col {
    	case "id":
    		recipients = append(recipients, &u.Id)
    	case "firstName":
    		recipients = append(recipients, &u.FirstName)
        case "lastName":
            recipients = append(recipients, &u.LastName)
        case "email":
            recipients = append(recipients, &u.Email)
        case "active":
            recipients = append(recipients, &u.Active)
        case "superuser":
            recipients = append(recipients, &u.Superuser)
        case "loginAttempts":
            recipients = append(recipients, &u.LoginAttempts)
        }
    }
    return recipients
}
