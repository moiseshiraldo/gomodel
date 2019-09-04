package gomodel

import (
	"fmt"
)

type ErrorTrace struct {
	App   *Application
	Model *Model
	Field string
	Err   error
}

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

type DatabaseError struct {
	Name  string
	Trace ErrorTrace
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("gomodel: %s db: %s", e.Name, e.Trace)
}

type ContainerError struct {
	Trace ErrorTrace
}

func (e *ContainerError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}

type QuerySetError struct {
	Trace ErrorTrace
}

func (e *QuerySetError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}

type ObjectNotFoundError struct {
	Trace ErrorTrace
}

func (e *ObjectNotFoundError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}

type MultipleObjectsError struct {
	Trace ErrorTrace
}

func (e *MultipleObjectsError) Error() string {
	return fmt.Sprintf("gomodel: %s", e.Trace)
}
