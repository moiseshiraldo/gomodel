package benchmark

import (
	"fmt"
	_ "github.com/gwenn/gosqlite"
	"github.com/moiseshiraldo/gomodel"
	"github.com/moiseshiraldo/gomodel/migration"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	app := gomodel.NewApp("main", "", User.Model)
	gomodel.Register(app)
	gomodel.Start(map[string]gomodel.Database{
		"default": {Driver: "sqlite3", Name: ":memory:"},
	})
	defer gomodel.Stop()
	err := migration.MakeAndRun("default")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	code := m.Run()
	User.Objects.All().Delete()
	os.Exit(code)
}
