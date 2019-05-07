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

func getConstructor(model Model) Constructor {
	switch model.meta.Constructor.(type) {
	case Values:
		return Values{}
	default:
		if builder, ok := model.meta.Constructor.(Builder); ok {
			return builder.New()
		} else {
			ct := reflect.TypeOf(model.meta.Constructor)
			if ct.Kind() == reflect.Ptr {
				ct = ct.Elem()
			}
			return reflect.New(ct)
		}
	}
}

func getConstructorType(constructor Constructor) string {
	switch constructor.(type) {
	case Values:
		return "Map"
	default:
		if _, ok := constructor.(Builder); ok {
			return "Builder"
		} else {
			ct := reflect.TypeOf(constructor)
			if ct.Kind() == reflect.Ptr {
				ct = ct.Elem()
			}
			if ct.Kind() == reflect.Struct {
				return "Struct"
			}
		}
		return ""
	}
}

func getRecipients(qs QuerySet, ct string) (Constructor, []interface{}) {
	var constructor Constructor
	recipients := make([]interface{}, 0, len(qs.Columns()))
	switch ct {
	case "Map":
		constructor = Values{}
		for _, name := range qs.Columns() {
			val := qs.Model().fields[name].NativeVal()
			recipients = append(recipients, &val)
		}
	case "Builder":
		builder := qs.Constructor().(Builder).New()
		recipients = builder.Recipients(qs.Columns())
		constructor = builder
	default:
		ct := reflect.TypeOf(qs.Constructor())
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
		constructor = cp.Interface()
	}
	return constructor, recipients
}
