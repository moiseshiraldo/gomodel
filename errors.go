package gomodel

import (
	"fmt"
)

// ErrorTrace holds the context information of a gomodel error.
type ErrorTrace struct {
	App   *Application
	Model *Model
	Field string
	Err   error
}

// String implements the fmt.Stringer interface.
func (e ErrorTrace) String() string {
	trace := ""
	if e.App != nil {
		trace += fmt.Sprintf("%s: ", e.App.name)
	}
	if e.Model != nil {
		trace += fmt.Sprintf("%s: ", e.Model.name)
	}
	if e.Field != "" {
		trace += fmt.Sprintf("%s: ", e.Field)
	}
	trace += e.Err.Error()
	return trace
}

// DatabaseError is raised when a database related error occurs.
type DatabaseError struct {
	Name  string // Name is the database key in the gomodel registry.
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("gomodel: %s db: %s", e.Name, e.Trace)
}

// ContainerError is raised when a model container related error occurs.
type ContainerError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *ContainerError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}

// QuerySetError is raised when a QuerySet related error occurs.
type QuerySetError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *QuerySetError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}

// ObjectNotFoundError is raised when the Get returns no objects.
type ObjectNotFoundError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *ObjectNotFoundError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}

// MultipleObjectsError is raised when the Get returns more than one object.
type MultipleObjectsError struct {
	Trace ErrorTrace
}

// Error implements the error interface.
func (e *MultipleObjectsError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}
