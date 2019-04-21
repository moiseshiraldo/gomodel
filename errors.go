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
		trace = trace + ": " + e.Model.name
	}
	if e.Field != "" {
		trace = trace + ": " + e.Field
	}
	if e.Err != nil {
		trace = trace + fmt.Sprintf(": %s", e.Err)
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
