package gomodels

import (
	"fmt"
	"testing"
)

// TestErrors tests gomodels error types
func TestErrors(t *testing.T) {
	app := &Application{name: "users"}
	model := &Model{name: "User"}
	t.Run("DatabaseError", func(t *testing.T) {
		err := &DatabaseError{
			"default",
			ErrorTrace{App: app, Model: model, Err: fmt.Errorf("test error")},
		}
		expected := "gomodels: default db: users: User: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("ContainerError", func(t *testing.T) {
		trace := ErrorTrace{
			App:   app,
			Model: model,
			Field: "email",
			Err:   fmt.Errorf("test error"),
		}
		err := &ContainerError{trace}
		expected := "gomodels: users: User: email: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("QuerySetError", func(t *testing.T) {
		trace := ErrorTrace{
			App:   app,
			Model: model,
			Err:   fmt.Errorf("test error"),
		}
		err := &QuerySetError{trace}
		expected := "gomodels: users: User: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("ObjectNotFoundError", func(t *testing.T) {
		trace := ErrorTrace{
			App:   app,
			Model: model,
			Err:   fmt.Errorf("object not found"),
		}
		err := &ObjectNotFoundError{trace}
		expected := "gomodels: users: User: object not found"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("MultipleObjectsError", func(t *testing.T) {
		trace := ErrorTrace{
			App:   app,
			Model: model,
			Err:   fmt.Errorf("multiple objects"),
		}
		err := &MultipleObjectsError{trace}
		expected := "gomodels: users: User: multiple objects"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
}
