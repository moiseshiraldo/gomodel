package benchmark

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodel"
	"os"
	"testing"
)

func loadMapQuerySet(b *testing.B) {
	os.Stdout, _ = os.Open(os.DevNull)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := User.Objects.Filter(
			gomodel.Q{"firstName": "Test"},
		).Load()
		if err != nil {
			b.Fatal(err)
		}
		for _, user := range users {
			fmt.Println(user.Display("email"))
		}
	}
}

func loadStructQuerySet(b *testing.B) {
	qs := User.Objects.WithContainer(userContainer{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := qs.Filter(gomodel.Q{"firstName": "Test"}).Load()
		if err != nil {
			b.Fatal(err)
		}
		for _, user := range users {
			fmt.Println(user.Display("email"))
		}
	}
}

func loadBuilderQuerySet(b *testing.B) {
	qs := User.Objects.WithContainer(&userBuilder{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users, err := qs.Filter(gomodel.Q{"firstName": "Test"}).Load()
		if err != nil {
			b.Fatal(err)
		}
		for _, user := range users {
			fmt.Println(user.Display("email"))
		}
	}
}

func loadRawSqlQuerySet(b *testing.B) {
	db := gomodel.Databases()["default"]
	for i := 0; i < b.N; i++ {
		query := `
            SELECT
              id, firstName, lastName, email, active, loginAttempts, created
            FROM
              "main_user"
            WHERE
              firstName = ?`
		rows, err := db.DB().Query(query, "Test")
		if err != nil {
			b.Fatal(err)
		}
		users := []*userContainer{}
		for rows.Next() {
			user := userContainer{}
			err = rows.Scan(
				&user.Id, &user.FirstName, &user.LastName, &user.Email,
				&user.Active, &user.LoginAttempts, &user.Created,
			)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, &user)
		}
		err = rows.Close()
		if err != nil {
			b.Fatal(err)
		}
		for _, user := range users {
			fmt.Println(user.Email)
		}
	}
}

func BenchmarkMultiRead(b *testing.B) {
	for i := 0; i < 100; i++ {
		_, err := User.Objects.Create(gomodel.Values{
			"firstName": "Test",
			"lastName":  "User",
			"email":     "user@test.com",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
	os.Stdout, _ = os.Open(os.DevNull)
	b.Run("RawSqlQuerySet", loadRawSqlQuerySet)
	b.Run("MapContainer", loadMapQuerySet)
	b.Run("StructContainer", loadStructQuerySet)
	b.Run("BuilderContainer", loadBuilderQuerySet)
	User.Objects.All().Delete()
}
