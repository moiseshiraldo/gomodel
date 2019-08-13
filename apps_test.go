package gomodels

import (
	"go/build"
	"path/filepath"
	"testing"
)

func TestApps(t *testing.T) {
	user := &Model{name: "User"}
	app := Application{
		name:   "users",
		path:   "tmp/migrations",
		models: map[string]*Model{"User": user},
	}
	t.Run("NewApp", func(t *testing.T) {
		settings := NewApp("test", "tmp/migrations", user)
		if settings.Name != "test" || settings.Path != "tmp/migrations" {
			t.Errorf("wrong app settings details")
		}
		if len(settings.Models) != 1 {
			t.Errorf("app settings missing model")
		}
	})
	t.Run("Name", func(t *testing.T) {
		if app.Name() != "users" {
			t.Errorf("expected users, got %s", app.Name())
		}
	})
	t.Run("Path", func(t *testing.T) {
		if app.Path() != "tmp/migrations" {
			t.Errorf("expected tmp/migrations, got %s", app.Path())
		}
	})
	t.Run("FullPath", func(t *testing.T) {
		path := filepath.Join(build.Default.GOPATH, "src", "tmp/migrations")
		if app.FullPath() != path {
			t.Errorf("expected %s, got %s", path, app.FullPath())
		}
	})
	t.Run("Models", func(t *testing.T) {
		if _, ok := app.Models()["User"]; !ok {
			t.Fatal("expected User model to be returned")
		}
		if app.Models()["User"] != user {
			t.Errorf("wrong user model")
		}
	})
}
