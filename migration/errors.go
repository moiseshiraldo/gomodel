package migration

import (
	"fmt"
)

// ErrorTrace holds the context information of a migration error.
type ErrorTrace struct {
	Node      *Node
	Operation Operation
	Err       error
}

// String implements the fmt.Stringer interface.
func (e ErrorTrace) String() string {
	trace := ""
	if e.Node != nil {
		trace += fmt.Sprintf("%s: %s: ", e.Node.App, e.Node.Name)
	}
	if e.Operation != nil {
		trace += fmt.Sprintf("%s: ", e.Operation.OpName())
	}
	trace += e.Err.Error()
	return trace
}

// AppNotFoundError is raised when the named application is not registered.
type AppNotFoundError struct {
	Name  string // Application name.
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *AppNotFoundError) Error() string {
	return fmt.Sprintf("migrations: %s: app not found", e.Name)
}

// NoAppMigrationsError is raised when trying to migrate an application with
// no migrations.
type NoAppMigrationsError struct {
	Name  string // Node name.
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *NoAppMigrationsError) Error() string {
	return fmt.Sprintf("migrations: %s: no migrations", e.Name)
}

// PathError is raised if reaading from the applicatoin path or writing to it
// is not possible.
type PathError struct {
	App   string // Application name.
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *PathError) Error() string {
	return fmt.Sprintf("migrations: %s: path error: %s", e.App, e.Trace)
}

// NameError is raised if an invalid migration node name is provided.
type NameError struct {
	Name  string // Node name.
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *NameError) Error() string {
	return fmt.Sprintf("migrations: %s: wrong node name", e.Name)
}

// DuplicateNumberError is raised if an application contains any migration
// file with a duplicate number.
type DuplicateNumberError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *DuplicateNumberError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

// LoadError is raised when a node fails to read from file.
type LoadError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *LoadError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

// SaveError is raised when a node fails to write to file.
type SaveError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *SaveError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

// InvalidDependencyError is raised when a migration file contains invalid
// dependencies.
type InvalidDependencyError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *InvalidDependencyError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

// CircularDependencyError is raised when the migration files contain circular
// dependencies.
type CircularDependencyError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

// OperationStateError is raised if a migration file contains invalid
// operations.
type OperationStateError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *OperationStateError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

// OperationRunError is raised if an operation fails to apply to the database
// schema.
type OperationRunError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *OperationRunError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}
