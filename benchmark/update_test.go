package benchmark

import (
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodel"
	"testing"
)

func updateMapContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(gomodel.Q{"firstName": "Test"}).Update(
			gomodel.Values{"firstName": "Updated", "email": "updated@test.com"},
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func updateStructContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(gomodel.Q{"firstName": "Test"}).Update(
			userContainer{FirstName: "Updated", Email: "updated@test.com"},
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func updateBuilderContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(gomodel.Q{"firstName": "Test"}).Update(
			userBuilder{FirstName: "Updated", Email: "updated@test.com"},
		)
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
              firstName = $1, email = $2
            WHERE
              firstName = $3`
		_, err := db.DB().Exec(
			query, "Updated", "updated@test.com", "Test",
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
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
	b.Run("RawSqlQuerySet", updateRawSqlContainer)
	b.Run("MapContainer", updateMapContainer)
	b.Run("StructContainer", updateStructContainer)
	b.Run("BuilderContainer", updateBuilderContainer)
	User.Objects.All().Delete()
}
