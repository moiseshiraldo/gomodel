package benchmarks

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodels"
	"os"
	"testing"
)

func insertMapContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
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
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
	}
}

func insertRawSqlContainer(b *testing.B) {
	db := gomodels.Databases()["default"]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := userContainer{
			FirstName: "Anakin",
			LastName:  "Skywalker",
			Email:     "anakin.skywalker@deathstar.com",
			Superuser: true,
		}
		query := `
            INSERT INTO
              "main_user" (firstName, lastName, email, superuser)
            VALUES
              ($1, $2, $3, $4)`
		result, err := db.Conn.Exec(
			query, user.FirstName, user.LastName, user.Email, user.Superuser,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		pk, err := result.LastInsertId()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		user.Id = int(pk)
	}
}

func BenchmarkInsert(b *testing.B) {
	b.Run("RawSqlContainer", insertRawSqlContainer)
	b.Run("MapContainer", insertMapContainer)
	b.Run("StructContainer", insertStructContainer)
	b.Run("BuilderContainer", insertBuilderContainer)
}
