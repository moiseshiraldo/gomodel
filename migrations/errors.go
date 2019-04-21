package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
)

type AppNotFoundError struct {
	Name string
	gomodels.ErrorTrace
}

func (e *AppNotFoundError) Error() string {
	return fmt.Sprintf("gomodels: migrations: %s: app not found", e.Name)
}

func (e *AppNotFoundError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type PathError struct {
	gomodels.ErrorTrace
}

func (e *PathError) Error() string {
	return fmt.Sprintf("migrations: load files: %s", e.ErrorTrace.String())
}

func (e *PathError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type NameError struct {
	Name string
	gomodels.ErrorTrace
}

func (e *NameError) Error() string {
	trace := e.ErrorTrace
	return fmt.Sprintf(
		"migrations: %s: %s", trace.App.Name(), e.Name,
	)
}

func (e *NameError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type DuplicateNumberError struct {
	Name string
	gomodels.ErrorTrace
}

func (e *DuplicateNumberError) Error() string {
	trace := e.ErrorTrace
	return fmt.Sprintf(
		"migrations: %s: duplicate number: %s", trace.App.Name(), e.Name,
	)
}

func (e *DuplicateNumberError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type LoadError struct {
	Name string
	gomodels.ErrorTrace
}

func (e *LoadError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: %s: load failed: %s",
		e.ErrorTrace.App.Name(), e.Name, e.ErrorTrace.String(),
	)
}

func (e *LoadError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type SaveError struct {
	Name string
	gomodels.ErrorTrace
}

func (e *SaveError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: %s: save failed: %s",
		e.ErrorTrace.App.Name(), e.Name, e.ErrorTrace.String(),
	)
}

func (e *SaveError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type InvalidDependencyError struct {
	Name       string
	Dependency string
	gomodels.ErrorTrace
}

func (e *InvalidDependencyError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: %s: invalid dependency: %s",
		e.ErrorTrace.App.Name(), e.Name, e.Dependency,
	)
}

func (e *InvalidDependencyError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type CircularDependencyError struct {
	Name       string
	Dependency string
	gomodels.ErrorTrace
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: %s: circular dependency: %s",
		e.ErrorTrace.App.Name(), e.Name, e.Dependency,
	)
}

func (e *CircularDependencyError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}

type OperationStateError struct {
	Node      string
	Operation *Operation
	gomodels.ErrorTrace
}

func (e *OperationStateError) Error() string {
	return fmt.Sprintf(
		"migrations: %s: %s: %s",
		e.ErrorTrace.App.Name(), e.Node, e.ErrorTrace.String(),
	)
}

func (e *OperationStateError) Trace() gomodels.ErrorTrace {
	return e.ErrorTrace
}
