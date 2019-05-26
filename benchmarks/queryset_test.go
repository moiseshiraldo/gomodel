package benchmarks

import (
    "testing"
    "github.com/moiseshiraldo/gomodels"
    "os"
    "fmt"
    _ "github.com/gwenn/gosqlite"
)

func loadMapQuerySet(b *testing.B) {
    os.Stdout,_ = os.Open(os.DevNull)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        users, _ := User.Objects.Filter(gomodels.Q{"firstName": "Luke"}).Load()
        for _, user := range users {
            fmt.Printf("%s", user.Get("email"))
        }
    }
}

func loadStructQuerySet(b *testing.B) {
    qs := User.Objects.SetContainer(userContainer{})
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        users, _ := qs.Filter(gomodels.Q{"firstName": "Luke"}).Load()
        for _, user := range users {
            fmt.Printf("%s", user.Get("email"))
        }
    }
}

func loadBuilderQuerySet(b *testing.B) {
    qs := User.Objects.SetContainer(&userBuilder{})
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        users, _ := qs.Filter(gomodels.Q{"firstName": "Luke"}).Load()
        for _, user := range users {
            fmt.Printf("%s", user.Get("email"))
        }
    }
}

func loadRawSqlQuerySet(b *testing.B) {
    db := gomodels.Databases["default"]
    for i := 0; i < b.N; i++ {
        query := `
            SELECT
              id, firstName, lastName, email, active, superuser, loginAttempts
            FROM
              'main_user'
            WHERE
              firstName = ?;`
        rows, _ := db.Query(query, "Luke")
        users := []*userContainer{}
        for rows.Next() {
            user := userContainer{}
            rows.Scan(
                &user.Id, &user.FirstName, &user.LastName, &user.Email,
                &user.Active, &user.Superuser, &user.LoginAttempts,
            )
            users = append(users, &user)
        }
        rows.Close()
        for _, user := range users {
            fmt.Printf("%s", user.Email)
        }
    }
}

func BenchmarkQuerySet(b *testing.B) {
    for i := 0; i < 1000; i++ {
        _, err := User.Objects.Create(gomodels.Values{
            "firstName": "Luke",
            "lastName": "Skywalker",
            "email": "luke.skywalker@deathstar.com",
            "superuser": true,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s", err)
            os.Exit(1)
        }
    }
    os.Stdout,_ = os.Open(os.DevNull)
    b.Run("RawSqlQuerySet", loadRawSqlQuerySet)
    b.Run("MapContainer", loadMapQuerySet)
    b.Run("StructContainer", loadStructQuerySet)
    b.Run("BuilderContainer", loadBuilderQuerySet)
}
