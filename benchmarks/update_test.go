package benchmarks

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodels"
	"os"
	"testing"
)

func updateMapContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(
			gomodels.Q{"firstName": "Anakin"},
		).Update(gomodels.Values{
			"firstName": "Darth",
			"lastName":  "Vader",
			"email":     "darth.vader@deathstar.com",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
	}
}

func updateStructContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(
			gomodels.Q{"firstName": "Anakin"},
		).Update(userContainer{
			FirstName: "Darth",
			LastName:  "Vader",
			Email:     "darth.vader@deathstar.com",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
	}
}

func updateBuilderContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := User.Objects.Filter(
			gomodels.Q{"firstName": "Anakin"},
		).Update(userBuilder{
			FirstName: "Darth",
			LastName:  "Vader",
			Email:     "darth.vader@deathstar.com",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
	}
}

func updateRawSqlContainer(b *testing.B) {
	db := gomodels.Databases()["default"]
	for i := 0; i < b.N; i++ {
		query := `
            UPDATE
              "main_user"
            SET
              firstName = $1, lastName = $2, email = $3
            WHERE
              firstName = $4`
		_, err := db.Conn.Exec(
			query, "Darth", "Vader", "darth.vader@deathstar.com", "Anakin",
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
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
	b.Run("RawSqlQuerySet", updateRawSqlContainer)
	b.Run("MapContainer", updateMapContainer)
	b.Run("StructContainer", updateStructContainer)
	b.Run("BuilderContainer", updateBuilderContainer)
	User.Objects.All().Delete()
}
