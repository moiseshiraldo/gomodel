package benchmark

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodel"
	"os"
	"testing"
)

func loadMapInstance(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := User.Objects.Get(gomodel.Q{"firstName": "Test"})
		if err != nil {
			b.Fatal(err)
		}
		fmt.Println(user.Display("email"))
	}
}

func loadStructInstance(b *testing.B) {
	qs := User.Objects.WithContainer(userContainer{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := qs.Get(gomodel.Q{"firstName": "Test"})
		if err != nil {
			b.Fatal(err)
		}
		fmt.Println(user.Display("email"))
	}
}

func loadBuilderInstance(b *testing.B) {
	qs := User.Objects.WithContainer(&userBuilder{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := qs.Get(gomodel.Q{"firstName": "Test"})
		if err != nil {
			b.Fatal(err)
		}
		fmt.Println(user.Display("email"))
	}
}

func loadRawSqlInstance(b *testing.B) {
	db := gomodel.Databases()["default"]
	for i := 0; i < b.N; i++ {
		user := userContainer{}
		query := `
            SELECT
              id, firstName, lastName, email, active, loginAttempts, created
            FROM
              "main_user"
            WHERE
              firstName = ?`
		err := db.DB().QueryRow(query, "Test").Scan(
			&user.Id, &user.FirstName, &user.LastName, &user.Email,
			&user.Active, &user.LoginAttempts, &user.Created,
		)
		if err != nil {
			b.Fatal(err)
		}
		fmt.Println(user.Email)
	}
}

func BenchmarkRead(b *testing.B) {
	_, err := User.Objects.Create(gomodel.Values{
		"firstName": "Test",
		"lastName":  "User",
		"email":     "user@test.com",
	})
	if err != nil {
		b.Fatal(err)
	}
	os.Stdout, _ = os.Open(os.DevNull)
	b.Run("RawSqlContainer", loadRawSqlInstance)
	b.Run("MapContainer", loadMapInstance)
	b.Run("StructContainer", loadStructInstance)
	b.Run("BuilderContainer", loadBuilderInstance)
	User.Objects.All().Delete()
}
