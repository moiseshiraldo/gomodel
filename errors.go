package gomodels

import (
	"fmt"
)

type Error interface {
	error
	Trace() ErrorTrace
}

type ErrorTrace struct {
	App   *Application
	Model *Model
	Field string
	Err   error
}

func (e *ErrorTrace) String() string {
	trace := ""
	if e.App != nil {
		trace += fmt.Sprintf(": %s", e.App.name)
	}
	if e.Model != nil {
		trace += fmt.Sprintf(": %s", e.Model.name)
	}
	if e.Field != "" {
		trace += fmt.Sprintf(": %s", e.Field)
	}
	if e.Err != nil {
		trace += fmt.Sprintf(": %s", e.Err)
	}
	return trace
}

type DatabaseError struct {
	Name string
	ErrorTrace
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("gomodels: %s", e.ErrorTrace.String())
}

func (e *DatabaseError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type ContainerError struct {
	ErrorTrace
}

func (e *ContainerError) Error() string {
	return fmt.Sprintf("gomodels: %s", e.ErrorTrace.String())
}

func (e *ContainerError) Trace() ErrorTrace {
	return e.ErrorTrace
}
