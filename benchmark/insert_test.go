package benchmark

import (
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodel"
	"testing"
	"time"
)

func insertMapContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Create(gomodel.Values{
			"firstName": "Test",
			"lastName":  "User",
			"email":     "user@test.com",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func insertStructContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Create(userContainer{
			FirstName: "Test",
			LastName:  "User",
			Email:     "user@test.com",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func insertBuilderContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Create(userBuilder{
			FirstName: "Test",
			LastName:  "User",
			Email:     "user@test.com",
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
			FirstName: "Test",
			LastName:  "User",
			Email:     "user@test.com",
			Active:    true,
			Created:   time.Now().Format("2006-01-02 15:04:05"),
		}
		query := `
            INSERT INTO
              "main_user" (
				  firstName, lastName, email, active, loginAttempts, created
			  )
            VALUES
              ($1, $2, $3, $4, $5, $6)`
		result, err := db.DB().Exec(
			query, user.FirstName, user.LastName, user.Email, user.Active,
			user.LoginAttempts, user.Created,
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
