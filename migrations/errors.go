package migrations

import "fmt"

type Error interface {
	error
	Trace() ErrorTrace
}

type ErrorTrace struct {
	Node      *Node
	Operation *Operation
	Err       error
}

func (e *ErrorTrace) String() string {
	trace := ""
	if e.Node != nil {
		trace += fmt.Sprintf("%s: %s", e.Node.App, e.Node.Name)
	}
	if (*e.Operation).OpName() != "" {
		trace += fmt.Sprintf(": %s", (*e.Operation).OpName())
	}
	if e.Err != nil {
		trace += fmt.Sprintf(": %s", e.Err)
	}
	return trace
}

type AppNotFoundError struct {
	Name string
	ErrorTrace
}

func (e *AppNotFoundError) Error() string {
	return fmt.Sprintf("migrations: %s: app not found", e.Name)
}

func (e *AppNotFoundError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type NoAppMigrationsError struct {
	Name string
	ErrorTrace
}

func (e *NoAppMigrationsError) Error() string {
	return fmt.Sprintf("migrations: %s: app not found", e.Name)
}

func (e *NoAppMigrationsError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type PathError struct {
	ErrorTrace
}

func (e *PathError) Error() string {
	return fmt.Sprintf("migrations: load files: %s", e.ErrorTrace.String())
}

func (e *PathError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type NameError struct {
	Name string
	ErrorTrace
}

func (e *NameError) Error() string {
	return fmt.Sprintf("migrations: %s: wrong name", e.Name)
}

func (e *NameError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type DuplicateNumberError struct {
	ErrorTrace
}

func (e *DuplicateNumberError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: duplicate number", e.ErrorTrace.String(),
	)
}

func (e *DuplicateNumberError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type LoadError struct {
	ErrorTrace
}

func (e *LoadError) Error() string {
	return fmt.Sprintf("migrations: %s: load failed", e.ErrorTrace.String())
}

func (e *LoadError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type SaveError struct {
	ErrorTrace
}

func (e *SaveError) Error() string {
	return fmt.Sprintf("migrations: %s: save failed", e.ErrorTrace.String())
}

func (e *SaveError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type InvalidDependencyError struct {
	ErrorTrace
}

func (e *InvalidDependencyError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: invalid dependency", e.ErrorTrace.String(),
	)
}

func (e *InvalidDependencyError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type CircularDependencyError struct {
	ErrorTrace
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: circular dependency", e.ErrorTrace.String(),
	)
}

func (e *CircularDependencyError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type OperationStateError struct {
	ErrorTrace
}

func (e *OperationStateError) Error() string {
	return fmt.Sprintf("migrations: state error: %s", e.ErrorTrace.String())
}

func (e *OperationStateError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type OperationRunError struct {
	ErrorTrace
}

func (e *OperationRunError) Error() string {
	return fmt.Sprintf("migrations: run error: %s:", e.ErrorTrace.String())
}

func (e *OperationRunError) Trace() ErrorTrace {
	return e.ErrorTrace
}
