package benchmarks

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodels"
	"os"
	"testing"
)

func loadMapQuerySet(b *testing.B) {
	os.Stdout, _ = os.Open(os.DevNull)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := User.Objects.Filter(
			gomodels.Q{"firstName": "Anakin"},
		).Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		for _, user := range users {
			fmt.Printf("%s", user.Get("email"))
		}
	}
}

func loadStructQuerySet(b *testing.B) {
	qs := User.Objects.SetContainer(userContainer{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := qs.Filter(gomodels.Q{"firstName": "Anakin"}).Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		for _, user := range users {
			fmt.Printf("%s", user.Get("email"))
		}
	}
}

func loadBuilderQuerySet(b *testing.B) {
	qs := User.Objects.SetContainer(&userBuilder{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := qs.Filter(gomodels.Q{"firstName": "Anakin"}).Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		for _, user := range users {
			fmt.Printf("%s", user.Get("email"))
		}
	}
}

func loadRawSqlQuerySet(b *testing.B) {
	db := gomodels.Databases()["default"]
	for i := 0; i < b.N; i++ {
		query := `
            SELECT
              id, firstName, lastName, email, active, superuser, loginAttempts
            FROM
              "main_user"
            WHERE
              firstName = ?`
		rows, err := db.Conn.Query(query, "Anakin")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		users := []*userContainer{}
		for rows.Next() {
			user := userContainer{}
			err = rows.Scan(
				&user.Id, &user.FirstName, &user.LastName, &user.Email,
				&user.Active, &user.Superuser, &user.LoginAttempts,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
				os.Exit(1)
			}
			users = append(users, &user)
		}
		err = rows.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		for _, user := range users {
			fmt.Printf("%s", user.Email)
		}
	}
}

func BenchmarkMultiRead(b *testing.B) {
	for i := 0; i < 100; i++ {
		_, err := User.Objects.Create(gomodels.Values{
			"firstName": "Anakin",
			"lastName":  "Skywalker",
			"email":     "anakin.skywalker@deathstar.com",
			"superuser": true,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
	}
	os.Stdout, _ = os.Open(os.DevNull)
	b.Run("RawSqlQuerySet", loadRawSqlQuerySet)
	b.Run("MapContainer", loadMapQuerySet)
	b.Run("StructContainer", loadStructQuerySet)
	b.Run("BuilderContainer", loadBuilderQuerySet)
}
