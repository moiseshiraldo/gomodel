package benchmark

import (
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodel"
	"testing"
)

func insertMapContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Create(gomodel.Values{
			"firstName": "Anakin",
			"lastName":  "Skywalker",
			"email":     "anakin.skywalker@deathstar.com",
			"superuser": true,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func insertStructContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Create(userContainer{
			FirstName: "Anakin",
			LastName:  "Skywalker",
			Email:     "anakin.skywalker@deathstar.com",
			Superuser: true,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func insertBuilderContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Create(userBuilder{
			FirstName: "Anakin",
			LastName:  "Skywalker",
			Email:     "anakin.skywalker@deathstar.com",
			Superuser: true,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func insertRawSqlContainer(b *testing.B) {
	db := gomodel.Databases()["default"]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := userContainer{
			FirstName:     "Anakin",
			LastName:      "Skywalker",
			Email:         "anakin.skywalker@deathstar.com",
			Superuser:     true,
			Active:        true,
			LoginAttempts: 0,
		}
		query := `
            INSERT INTO
              "main_user" (
				  firstName, lastName, email, superuser, active, loginAttempts
			  )
            VALUES
              ($1, $2, $3, $4, $5, $6)`
		result, err := db.Conn().Exec(
			query, user.FirstName, user.LastName, user.Email, user.Superuser,
			user.Active, user.LoginAttempts,
		)
		if err != nil {
			b.Fatal(err)
		}
		pk, err := result.LastInsertId()
		if err != nil {
			b.Fatal(err)
		}
		user.Id = int(pk)
	}
}

func BenchmarkInsert(b *testing.B) {
	b.Run("RawSqlContainer", insertRawSqlContainer)
	b.Run("MapContainer", insertMapContainer)
	b.Run("StructContainer", insertStructContainer)
	b.Run("BuilderContainer", insertBuilderContainer)
	User.Objects.All().Delete()
}
