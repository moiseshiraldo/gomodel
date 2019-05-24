package gomodels

import (
	"fmt"
	"reflect"
	"strings"
)

func sqlColumnOptions(null bool, pk bool, unique bool) string {
	options := ""
	if null {
		options += " NULL"
	} else {
		options += " NOT NULL"
	}
	if pk {
		options += " PRIMARY KEY"
	} else if unique {
		options += " UNIQUE"
	}
	return options
}

func sqlCreateQuery(table string, values Values) (string, []interface{}) {
	cols := make([]string, 0, len(values))
	vals := make([]interface{}, 0, len(values))
	phs := make([]string, 0, len(values))
	index := 1
	for col, val := range values {
		cols = append(cols, fmt.Sprintf("'%s'", col))
		vals = append(vals, val)
		phs = append(phs, fmt.Sprintf("$%d", index))
		index += 1
	}
	colStr := strings.Join(cols, ", ")
	phStr := strings.Join(phs, ", ")
	query := fmt.Sprintf(
		"INSERT INTO '%s' (%s) VALUES (%s)", table, colStr, phStr,
	)
	return query, vals
}

func getContainerType(container Container) string {
	switch container.(type) {
	case Values:
		return containers.Map
	default:
		if _, ok := container.(Builder); ok {
			return containers.Builder
		} else {
			ct := reflect.TypeOf(container)
			if ct.Kind() == reflect.Ptr {
				ct = ct.Elem()
			}
			if ct.Kind() == reflect.Struct {
				return containers.Struct
			}
		}
		return ""
	}
}

func getRecipients(qs QuerySet, conType string) (Container, []interface{}) {
	var container Container
	recipients := make([]interface{}, 0, len(qs.Columns()))
	switch conType {
	case containers.Map:
		container = Values{}
		for _, name := range qs.Columns() {
			val := qs.Model().fields[name].NativeVal()
			recipients = append(recipients, &val)
		}
	case containers.Builder:
		builder := qs.Container().(Builder).New()
		recipients = builder.Recipients(qs.Columns())
		container = builder
	default:
		ct := reflect.TypeOf(qs.Container())
		if ct.Kind() == reflect.Ptr {
			ct = ct.Elem()
		}
		cp := reflect.New(ct)
		for _, name := range qs.Columns() {
			f := cp.Elem().FieldByName(strings.Title(name))
			if f.IsValid() && f.CanAddr() {
				recipients = append(recipients, f.Addr().Interface())
			}
		}
		container = cp.Interface()
	}
	return container, recipients
}
