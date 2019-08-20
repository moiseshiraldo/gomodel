package benchmarks

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodels"
	"os"
	"testing"
)

func loadMapInstance(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := User.Objects.Get(gomodels.Q{"firstName": "Anakin"})
		if err != nil {
			b.Fatal(err)
		}
		fmt.Printf("%s", user.Get("email"))
	}
}

func loadStructInstance(b *testing.B) {
	qs := User.Objects.SetContainer(userContainer{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := qs.Get(gomodels.Q{"firstName": "Anakin"})
		if err != nil {
			b.Fatal(err)
		}
		fmt.Printf("%s", user.Get("email"))
	}
}

func loadBuilderInstance(b *testing.B) {
	qs := User.Objects.SetContainer(&userBuilder{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := qs.Get(gomodels.Q{"firstName": "Anakin"})
		if err != nil {
			b.Fatal(err)
		}
		fmt.Printf("%s", user.Get("email"))
	}
}

func loadRawSqlInstance(b *testing.B) {
	db := gomodels.Databases()["default"]
	for i := 0; i < b.N; i++ {
		user := userContainer{}
		query := `
            SELECT
              id, firstName, lastName, email, active, superuser, loginAttempts
            FROM
              "main_user"
            WHERE
              firstName = ?`
		err := db.Conn().QueryRow(query, "Anakin").Scan(
			&user.Id, &user.FirstName, &user.LastName, &user.Email,
			&user.Active, &user.Superuser, &user.LoginAttempts,
		)
		if err != nil {
			b.Fatal(err)
		}
		fmt.Printf("%s", user.Email)
	}
}

func BenchmarkRead(b *testing.B) {
	_, err := User.Objects.Create(gomodels.Values{
		"firstName": "Anakin",
		"lastName":  "Skywalker",
		"email":     "anakin.skywalker@deathstar.com",
		"superuser": true,
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
