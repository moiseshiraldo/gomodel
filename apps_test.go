package gomodels

import (
	"go/build"
	"path/filepath"
	"testing"
)

// TestApp tests the Application struct methods
func TestApp(t *testing.T) {
	user := &Model{name: "User"}
	app := &Application{
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

	t.Run("RelativePath", func(t *testing.T) {
		path := filepath.Join(build.Default.GOPATH, "src", "tmp/migrations")
		if app.FullPath() != path {
			t.Errorf("expected %s, got %s", path, app.FullPath())
		}
	})

	t.Run("FullPath", func(t *testing.T) {
		app.path = "/tmp/migrations"
		if app.FullPath() != app.path {
			t.Errorf("expected %s, got %s", app.path, app.FullPath())
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

// TestRegistry tests app registry related functions
func TestRegistry(t *testing.T) {
	registry["users"] = &Application{name: "users"}
	defer ClearRegistry()

	t.Run("Get", func(t *testing.T) {
		reg := Registry()
		if _, ok := reg["users"]; !ok {
			t.Error("returned registry is missing app users")
		}
	})

	t.Run("Modify", func(t *testing.T) {
		reg := Registry()
		delete(reg, "users")
		if _, ok := registry["users"]; !ok {
			t.Error("app users was deleted from internal registry")
		}
	})

	t.Run("AddDuplicateApp", func(t *testing.T) {
		registry["customers"] = &Application{name: "customers"}
		appSettings := AppSettings{Name: "customers"}
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Register function to panic")
			}
			delete(registry, "customers")
		}()
		Register(appSettings)
	})

	t.Run("AddAppInvalidModel", func(t *testing.T) {
		customer := &Model{
			name:   "Customer",
			fields: Fields{},
			meta:   Options{Container: "Invalid container"},
		}
		appSettings := AppSettings{
			Name:   "customers",
			Models: []*Model{customer},
		}
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Register function to panic")
			}
			delete(registry, "customers")
		}()
		Register(appSettings)
	})

	t.Run("AddApp", func(t *testing.T) {
		customer := &Model{name: "Customer", fields: Fields{}}
		appSettings := AppSettings{
			Name:   "customers",
			Models: []*Model{customer},
		}
		Register(appSettings)
		if _, ok := registry["customers"]; !ok {
			t.Fatal("internal registry is missing app customers")
		}
		app := registry["customers"]
		if _, ok := app.models["Customer"]; !ok {
			t.Fatal("app is missing Customer model")
		}
		if app.models["Customer"].app != app {
			t.Fatal("expected model to be linked to app")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		registry["services"] = &Application{name: "services"}
		ClearRegistry()
		if len(registry) > 0 {
			t.Errorf("app registry was not cleared")
		}
	})
}
