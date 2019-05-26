package benchmarks

import (
    "testing"
    "github.com/moiseshiraldo/gomodels"
    "github.com/moiseshiraldo/gomodels/migrations"
    "os"
    "fmt"
    _ "github.com/gwenn/gosqlite"
)

func TestMain(m *testing.M) {
    app := gomodels.NewApp("main", "", User.Model)
    gomodels.Register(app)
    gomodels.Start(map[string]gomodels.Database{
        "default": gomodels.Database{Driver: "sqlite3", Name: ":memory:"},
    })
    defer gomodels.Stop()
    err := migrations.MakeAndRun()
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s", err)
        os.Exit(1)
    }
    code := m.Run()
    User.Objects.All().Delete()
    os.Exit(code)
}
