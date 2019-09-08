package benchmark

import (
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodel"
	"testing"
)

func updateMapContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(
			gomodel.Q{"firstName": "Anakin"},
		).Update(gomodel.Values{
			"firstName": "Darth",
			"lastName":  "Vader",
			"email":     "darth.vader@deathstar.com",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func updateStructContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(
			gomodel.Q{"firstName": "Anakin"},
		).Update(userContainer{
			FirstName: "Darth",
			LastName:  "Vader",
			Email:     "darth.vader@deathstar.com",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func updateBuilderContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(
			gomodel.Q{"firstName": "Anakin"},
		).Update(userBuilder{
			FirstName: "Darth",
			LastName:  "Vader",
			Email:     "darth.vader@deathstar.com",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func updateRawSqlContainer(b *testing.B) {
	db := gomodel.Databases()["default"]
	for i := 0; i < b.N; i++ {
		query := `
            UPDATE
              "main_user"
            SET
              firstName = $1, lastName = $2, email = $3
            WHERE
              firstName = $4`
		_, err := db.DB().Exec(
			query, "Darth", "Vader", "darth.vader@deathstar.com", "Anakin",
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	for i := 0; i < 100; i++ {
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
	b.Run("RawSqlQuerySet", updateRawSqlContainer)
	b.Run("MapContainer", updateMapContainer)
	b.Run("StructContainer", updateStructContainer)
	b.Run("BuilderContainer", updateBuilderContainer)
	User.Objects.All().Delete()
}
