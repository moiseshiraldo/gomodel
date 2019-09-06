package migration

import (
	"fmt"
	"testing"
)

// TestErrors tests migrations error types
func TestErrors(t *testing.T) {
	node := &Node{App: "users", number: 1, name: "initial"}
	operation := &mockedOperation{}
	t.Run("AppNotFoundError", func(t *testing.T) {
		err := &AppNotFoundError{"users", ErrorTrace{}}
		expected := "migrations: users: app not found"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("NoAppMigrationsError", func(t *testing.T) {
		err := &NoAppMigrationsError{"users", ErrorTrace{}}
		expected := "migrations: users: no migrations"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("PathError", func(t *testing.T) {
		err := &PathError{"users", ErrorTrace{Err: fmt.Errorf("test error")}}
		expected := "migrations: users: path error: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("NameError", func(t *testing.T) {
		err := &NameError{"qwerty", ErrorTrace{}}
		expected := "migrations: qwerty: wrong node name"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("DuplicateNumberError", func(t *testing.T) {
		err := &DuplicateNumberError{
			ErrorTrace{Node: node, Err: fmt.Errorf("duplicate number")},
		}
		expected := "migrations: users: 0001_initial: duplicate number"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("LoadError", func(t *testing.T) {
		err := &LoadError{ErrorTrace{Node: node, Err: fmt.Errorf("test error")}}
		expected := "migrations: users: 0001_initial: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("SaveError", func(t *testing.T) {
		err := &SaveError{ErrorTrace{Node: node, Err: fmt.Errorf("test error")}}
		expected := "migrations: users: 0001_initial: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("InvalidDependencyError", func(t *testing.T) {
		err := &InvalidDependencyError{
			ErrorTrace{Node: node, Err: fmt.Errorf("invalid dependency")}}
		expected := "migrations: users: 0001_initial: invalid dependency"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("CircularDependencyError", func(t *testing.T) {
		err := &CircularDependencyError{
			ErrorTrace{Node: node, Err: fmt.Errorf("circular dependency")}}
		expected := "migrations: users: 0001_initial: circular dependency"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("OperationStateError", func(t *testing.T) {
		err := &OperationStateError{
			ErrorTrace{node, operation, fmt.Errorf("test error")},
		}
		expected := `migrations: users: 0001_initial: ` +
			`MockedOperation: test error`
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
	t.Run("OperationRunError", func(t *testing.T) {
		err := &OperationRunError{
			ErrorTrace{node, operation, fmt.Errorf("test error")},
		}
		expected := `migrations: users: 0001_initial: ` +
			`MockedOperation: test error`
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
}
