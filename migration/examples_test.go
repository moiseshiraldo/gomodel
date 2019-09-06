package migration

import (
	"fmt"
	"github.com/moiseshiraldo/gomodel"
)

func ExampleMake() {
	User := gomodel.New(
		"User",
		gomodel.Fields{
			"email":   gomodel.CharField{Unique: true, MaxLength: 100},
			"active":  gomodel.BooleanField{DefaultFalse: true},
			"created": gomodel.DateTimeField{AutoNowAdd: true},
		},
		gomodel.Options{},
	)
	app := gomodel.NewApp("users", "myproject/users/migrations", User.Model)
	gomodel.Register(app)
	gomodel.Start(map[string]gomodel.Database{
		"default": {
			Driver:   "postgres",
			Name:     "gomodeltest",
			User:     "local",
			Password: "1234",
		},
	})
	if _, err := Make("users", MakeOptions{}); err != nil {
		fmt.Println(err)
	}
}

func ExampleRun() {
	User := gomodel.New(
		"User",
		gomodel.Fields{
			"email":   gomodel.CharField{Unique: true, MaxLength: 100},
			"active":  gomodel.BooleanField{DefaultFalse: true},
			"created": gomodel.DateTimeField{AutoNowAdd: true},
		},
		gomodel.Options{},
	)
	app := gomodel.NewApp("users", "myproject/users/migrations", User.Model)
	gomodel.Register(app)
	gomodel.Start(map[string]gomodel.Database{
		"default": {
			Driver:   "postgres",
			Name:     "gomodeltest",
			User:     "local",
			Password: "1234",
		},
	})
	if err := Run(RunOptions{App: "users", Node: "0001_inital"}); err != nil {
		fmt.Println(err)
	}
}

func EampleMakeAndRun() {
	Player := gomodel.New(
		"Player",
		gomodel.Fields{
			"username": gomodel.CharField{Unique: true, MaxLength: 100},
			"score":    gomodel.IntegerField{DefaultZero: true},
		},
		gomodel.Options{},
	)
	app := gomodel.NewApp("main", "", Player.Model)
	gomodel.Register(app)
	gomodel.Start(map[string]gomodel.Database{
		"default": {
			Driver: "sqlite3",
			Name:   ":memory:",
		},
	})
	if err := MakeAndRun("default"); err != nil {
		fmt.Println(err)
	}
}
