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
	trace := e.App.name
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

type DuplicateAppError struct {
	ErrorTrace
}

func (e *DuplicateAppError) Error() string {
	return fmt.Sprintf("gomodels: %s: duplicate app", e.ErrorTrace.String())
}

func (e *DuplicateAppError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type DuplicateModelError struct {
	ErrorTrace
}

func (e *DuplicateModelError) Error() string {
	return fmt.Sprintf("gomodels: %s: duplicate model", e.ErrorTrace.String())
}

func (e *DuplicateModelError) Trace() ErrorTrace {
	return e.ErrorTrace
}

type DuplicatePkError struct {
	ErrorTrace
}

func (e *DuplicatePkError) Error() string {
	return fmt.Sprintf("gomodels: %s: duplicate pk", e.ErrorTrace.String())
}

func (e *DuplicatePkError) Trace() ErrorTrace {
	return e.ErrorTrace
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

type ConstructorError struct {
	ErrorTrace
}

func (e *ConstructorError) Error() string {
	return fmt.Sprintf("gomodels: %s", e.ErrorTrace.String())
}

func (e *ConstructorError) Trace() ErrorTrace {
	return e.ErrorTrace
}
