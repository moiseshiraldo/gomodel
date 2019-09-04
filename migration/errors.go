package migration

import "fmt"

type ErrorTrace struct {
	Node      *Node
	Operation Operation
	Err       error
}

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

type AppNotFoundError struct {
	Name  string
	Trace ErrorTrace
}

func (e *AppNotFoundError) Error() string {
	return fmt.Sprintf("migrations: %s: app not found", e.Name)
}

type NoAppMigrationsError struct {
	Name  string
	Trace ErrorTrace
}

func (e *NoAppMigrationsError) Error() string {
	return fmt.Sprintf("migrations: %s: no migrations", e.Name)
}

type PathError struct {
	App   string
	Trace ErrorTrace
}

func (e *PathError) Error() string {
	return fmt.Sprintf("migrations: %s: path error: %s", e.App, e.Trace)
}

type NameError struct {
	Name  string
	Trace ErrorTrace
}

func (e *NameError) Error() string {
	return fmt.Sprintf("migrations: %s: wrong node name", e.Name)
}

type DuplicateNumberError struct {
	Trace ErrorTrace
}

func (e *DuplicateNumberError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

type LoadError struct {
	Trace ErrorTrace
}

func (e *LoadError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

type SaveError struct {
	Trace ErrorTrace
}

func (e *SaveError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

type InvalidDependencyError struct {
	Trace ErrorTrace
}

func (e *InvalidDependencyError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

type CircularDependencyError struct {
	Trace ErrorTrace
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

type OperationStateError struct {
	Trace ErrorTrace
}

func (e *OperationStateError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}

type OperationRunError struct {
	Trace ErrorTrace
}

func (e *OperationRunError) Error() string {
	return fmt.Sprintf("migrations: %s", e.Trace)
}
